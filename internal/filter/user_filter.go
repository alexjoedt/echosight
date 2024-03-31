package filter

type UserFilter struct {
	Filter
	ID        string
	FirstName string
	LastName  string
	Role      string
	Email     string
	Active    bool
}

func NewDefaultUserFilter() *UserFilter {
	return &UserFilter{
		Filter: NewDefaultFilter(),
	}
}
