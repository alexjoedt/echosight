package filter

import "github.com/google/uuid"

type DetectorFilter struct {
	Filter
	Name   *string
	HostID *uuid.UUID
	Type   *string
	Active *bool
}

func NewDefaultDetectorFilter() *DetectorFilter {
	return &DetectorFilter{
		Filter: NewDefaultFilter(),
	}
}
