package echosight

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// User represents the user for this application
type Recipient struct {
	bun.BaseModel `bun:"table:recipients"`
	ID            uuid.UUID `json:"id" bun:"type:uuid,pk,default:uuid_generate_v4()"`
	LookupVersion int       `json:"lookupVersion" bun:",default:1"`
	Name          string    `json:"first_name"`
	Activated     bool      `json:"activated"`
	Email         string    `json:"email"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
