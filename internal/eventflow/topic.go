package eventflow

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/rs/xid"
)

type Topic struct {
	ID string

	mu            sync.RWMutex
	subscriptions map[string]*Subscription

	engine *Engine
}

// Subscribe subcribes to the topic.
// Starts a go routine in the background.
func (t *Topic) Subscribe(ctx context.Context, onEventFn EventHandler) (*Subscription, error) {
	sub, err := t.newSubscription()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event, ok := <-sub.event:
				if !ok {
					t.engine.log.Debugw("subscription closed", logger.Str("topic_id", t.ID))
					return
				}

				if err := onEventFn(ctx, event); err != nil {
					t.engine.log.Errorf("failed to process event: '%v'", err)
				}

			case <-sub.done:
				close(sub.event)
				return

			case <-ctx.Done():
				t.engine.log.Debugw("context done", logger.Str("topic_id", sub.topic.ID), logger.Str("sub_id", sub.ID))
				close(sub.event)
				return
			}
		}
	}()

	return sub, nil

}

func (t *Topic) SubscribeChannel() (chan *Event, error) {
	sub, err := t.newSubscription()
	if err != nil {
		return nil, err
	}

	return sub.event, nil
}

func (t *Topic) Publish(event *Event) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, s := range t.subscriptions {
		event.topicID = t.ID
		event.subscriptionID = s.ID
		s.event <- event
	}

	if len(t.subscriptions) != 0 {
		t.engine.log.Debugw("event published", logger.Str("topic_id", t.ID))
	}
}

func (t *Topic) Close() error {
	t.mu.Lock()
	for _, s := range t.subscriptions {
		s.Close()
	}
	t.mu.Unlock()

	t.engine.mu.Lock()
	delete(t.engine.topics, t.ID)
	t.engine.mu.Unlock()
	return nil
}

func (t *Topic) newSubscription() (*Subscription, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := xid.New().String()
	t.subscriptions[id] = &Subscription{
		ID:    id,
		event: make(chan *Event, 1),
		done:  make(chan struct{}, 1),
		topic: t,
	}

	return t.subscriptions[id], nil
}

func (t *Topic) removeSubscription(subID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	sub, ok := t.subscriptions[subID]
	if !ok {
		return fmt.Errorf("no subscription with this id on topic")
	}
	sub.done <- struct{}{}
	delete(t.subscriptions, subID)
	return nil
}
