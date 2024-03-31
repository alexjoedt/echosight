package echosight

type Role string

func (r Role) String() string {
	return string(r)
}

const (
	RoleAdmin   Role = "admin"
	RoleRegular Role = "regular"
)
