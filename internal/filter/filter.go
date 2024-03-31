package filter

import (
	"fmt"
	"math"
	"strings"

	"github.com/alexjoedt/echosight/internal/validator"
)

type Filter struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
	Pagination   Pagination
}

func NewDefaultFilter() Filter {
	return Filter{
		Page:         1,
		PageSize:     100_000,
		Sort:         "created_at",
		SortSafelist: []string{"created_at", "-created_at"},
	}
}

func ValidateFilters(v *validator.Validator, f Filter) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 10_000_000, "page_size", "must be a maximum of 100")
	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

func (f Filter) SortColumn() string {
	for _, sv := range f.SortSafelist {
		if f.Sort == sv {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	return ""
}

func (f Filter) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filter) Order() string {
	return fmt.Sprintf("%s %s", f.SortColumn(), f.SortDirection())
}

func (f Filter) Limit() int {
	if f.PageSize == 0 {
		return 25
	}
	return f.PageSize
}

func (f Filter) Offset() int {
	return (f.Page - 1) * f.PageSize
}

type Pagination struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func ComputePagination(totalRecords int, page int, pageSize int) Pagination {
	if totalRecords == 0 {
		return Pagination{}
	}

	return Pagination{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
