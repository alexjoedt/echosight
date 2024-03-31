package filter

type PreferenceFilter struct {
	Filter
	Name  string
	Value string
}

func NewDefaultPreferenceFilter() *PreferenceFilter {
	return &PreferenceFilter{
		Filter: NewDefaultFilter(),
	}
}
