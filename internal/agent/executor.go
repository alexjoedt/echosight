package agent

import (
	"context"
)

type Executor interface {
	Execute(ctx context.Context) (*Result, error)
}
