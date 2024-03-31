package echosight

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:sessions"`

	Token  string    `json:"token" bun:"-"`
	Hash   []byte    `json:"-" bun:"hash,pk"`
	UserID uuid.UUID `json:"-"`
	Expiry time.Time `json:"expiry" bun:"expiry"`

	User *User `json:"-" bun:"rel:belongs-to,join:user_id=id"`
}
