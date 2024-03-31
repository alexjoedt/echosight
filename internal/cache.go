package echosight

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, namespace string, id string) ([]byte, error)
	Put(ctx context.Context, namespace string, id string, data []byte, duration time.Duration) error
	Update(ctx context.Context, namespace string, id string, data []byte, duration time.Duration) error
	Delete(ctx context.Context, namespace string, id string) error
}
