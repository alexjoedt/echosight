package echosight

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	UserSessionKey string = "userID"
)

type Users []User

var AnonymusUser = &User{}

// User represents the user for this application
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`
	ID            uuid.UUID `json:"id" bun:"type:uuid,pk,default:uuid_generate_v4()"`
	LookupVersion int       `json:"lookupVersion" bun:",default:1"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Activated     bool      `json:"activated"`
	Email         string    `json:"email"`
	Password      Password  `json:"-" bun:"-"`
	PasswordHash  []byte    `json:"-"`
	Role          Role      `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     time.Time `json:"deleted_at"`
	Image         string    `json:"image" bun:"-"`

	// Preferences UserPreferences `json:"preferences" bson:"preferences"`
}

// PublicUser represents a public user for this application
// with limited information
type PublicUser struct {
	ID        uuid.UUID `json:"id" bun:"type:uuid,pk,default:uuid_generate_v4()"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Activated bool      `json:"activated"`
	CreatedAt time.Time `json:"created_at"`
}

func (user *User) Public() *PublicUser {
	userJSON, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	var publicUser PublicUser
	json.Unmarshal(userJSON, &publicUser)
	return &publicUser
}

func (user *User) IsAdmin() bool {
	return user.Role == RoleAdmin
}

func (user *User) IsAnonymus() bool {
	return user == AnonymusUser
}
