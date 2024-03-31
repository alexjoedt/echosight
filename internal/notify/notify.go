package notify

import (
	"context"
	"errors"
	"fmt"
	"sync"

	es "github.com/alexjoedt/echosight/internal"
)

type Sender interface {
	Send(ctx context.Context, result *es.Result) error
	Enabled() bool
}

type Notifier struct {
	mu       sync.RWMutex
	registry map[string]Sender
}

func NewNotifier() *Notifier {
	return &Notifier{
		mu:       sync.RWMutex{},
		registry: make(map[string]Sender, 0),
	}
}

func (n *Notifier) AddSender(id string, sender Sender) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.registry[id]; ok {
		return fmt.Errorf("sender with id '%s' already registered", id)
	}
	n.registry[id] = sender
	return nil
}

// Send sends the result to all registered sender
func (n *Notifier) Send(ctx context.Context, result *es.Result) error {
	var sendErrors []error
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, s := range n.registry {
		if s.Enabled() {
			if err := s.Send(ctx, result); err != nil {
				sendErrors = append(sendErrors, err)
			}
		}
	}

	if len(sendErrors) > 1 {
		return errors.Join(sendErrors...)
	}
	return nil
}
