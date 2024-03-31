package echosight

import (
	"context"

	"github.com/google/uuid"
)

type contextKey int

const (
	userIDContextKey = contextKey(iota + 1)
	userContextKey
	correlationKey
)

func NewContextWithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationKey, id)
}

func CorrelationIDFromContext(ctx context.Context) string {
	id, ok := ctx.Value(correlationKey).(string)
	if ok {
		return id
	}
	return ""
}

// NewContextWithUser returns a new context with the given user.
func NewContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext returns the current logged in user.
func UserFromContext(ctx context.Context) (*User, error) {
	ctxVal := ctx.Value(userContextKey)
	if ctxVal == nil {
		return nil, ErrInternalf("no user in context")
	}
	user, ok := ctxVal.(*User)
	if !ok {
		return nil, ErrInvalidf("invalid user in contexz")
	}

	return user, nil
}

// UserIDFromContext is a helper function that returns the ID of the current
// logged in user. Returns an empty string if no user is logged in.
func UserIDFromContext(ctx context.Context) uuid.UUID {
	if user, _ := UserFromContext(ctx); user != nil {
		return user.ID
	}
	return uuid.Nil
}
