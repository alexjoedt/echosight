package filter

import "github.com/google/uuid"

type RecipientFilter struct {
	Filter
	HostID *uuid.UUID
	Name   *string
	Email  *string
	Active *bool
}

func NewDefaultRecipientFilter() *RecipientFilter {
	return &RecipientFilter{
		Filter: NewDefaultFilter(),
	}
}
