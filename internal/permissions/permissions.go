package permissions

import (
	"fmt"
	"strings"
)

type Permission int

const (
	P_READ Permission = 1 << iota
	P_WRITE
	P_DELETE
)

func (p *Permission) Read() bool {
	return *p&P_READ != 0
}

func (p *Permission) Write() bool {
	return *p&P_WRITE != 0
}

func (p *Permission) Delete() bool {
	return *p&P_DELETE != 0
}

func (p *Permission) Set(t Permission) {
	// write permission makes no sense without read
	if t.Write() && !t.Read() {
		t = t | P_READ
	}

	// delete permission makes no sense without read and write
	if t.Delete() && !t.Write() {
		t = t | P_WRITE
	}

	if t.Delete() && !t.Read() {
		t = t | P_READ
	}

	*p = t
}

func (p *Permission) Clear(t Permission) {
	*p = *p &^ t
}

func (p *Permission) String() string {
	switch *p {
	case P_READ:
		return "r"
	case P_WRITE:
		return "w"
	case P_DELETE:
		return "d"
	case P_READ | P_WRITE:
		return "rw"
	case P_READ | P_WRITE | P_DELETE:
		return "rwd"
	case P_READ | P_DELETE:
		return "rd" //
	case P_WRITE | P_DELETE:
		return "wd"
	default:
		return ""
	}
}

var _ fmt.Stringer = (*Permission)(nil)

// FromString returns permissions from 'rwd', 'r' string
func FromString(s string) Permission {
	parts := strings.Split(s, "")

	var perms Permission
	for _, p := range parts {
		switch p {
		case "r":
			perms = perms | P_READ
		case "w":
			perms = perms | P_WRITE
		case "d":
			perms = perms | P_DELETE
		}
	}

	return perms
}
