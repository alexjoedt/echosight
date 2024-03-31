package filter

type HostFilter struct {
	Filter
	Name      string
	DNS       string
	IPv4      string
	IPv6      string
	Location  string
	Activated bool
}

func NewDefaultHostFilter() *HostFilter {
	return &HostFilter{
		Filter: NewDefaultFilter(),
	}
}
