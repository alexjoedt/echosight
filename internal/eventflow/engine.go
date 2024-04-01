package eventflow

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexjoedt/echosight/internal/logger"
)

type eventWrapper struct {
	topicID string
	event   *Event
}

type Engine struct {
	mu     sync.RWMutex
	topics map[string]*Topic
	log    *logger.Logger
}

func NewEngine() *Engine {
	e := &Engine{
		topics: make(map[string]*Topic, 0),
		log:    logger.New("EventFlow"),
	}

	return e
}

// NewTopic, creates a new topic on the engine
func (e *Engine) NewTopic(id string) (*Topic, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.topics[id]; ok {
		return nil, fmt.Errorf("topic with id '%s' already exists", id)
	}

	t := &Topic{
		ID:            id,
		subscriptions: make(map[string]*Subscription),
		engine:        e,
	}

	e.topics[id] = t
	e.log.Debugw("topic created", logger.Str("topic_id", t.ID))
	return t, nil
}

func (e *Engine) CloseTopic(id string) error {
	t, err := e.GetTopic(id)
	if err != nil {
		return err
	}
	return t.Close()
}

func (e *Engine) GetTopic(id string) (*Topic, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if t, ok := e.topics[id]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("topic with id '%s' does not exists", id)
}

func (e *Engine) Publish(topicID string, event *Event) error {
	t, err := e.GetTopic(topicID)
	if err != nil {
		return err
	}

	t.Publish(event)
	return nil
}

// Subscribe subcribes to a topic.
// Starts a go routine in the background.
func (e *Engine) Subscribe(ctx context.Context, topicID string, onEventFn EventHandler) (*Subscription, error) {
	t, err := e.GetTopic(topicID)
	if err != nil {
		return nil, err
	}

	return t.Subscribe(ctx, onEventFn)
}

func (e *Engine) SubscribeChannel(topicID string) (chan *Event, error) {
	t, err := e.GetTopic(topicID)
	if err != nil {
		return nil, err
	}
	return t.SubscribeChannel()
}

func (e *Engine) CloseSubscription(topic string, subID string) error {
	t, err := e.GetTopic(topic)
	if err != nil {
		return err
	}
	return t.removeSubscription(subID)
}

func (e *Engine) Stop() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, t := range e.topics {
		for _, s := range t.subscriptions {
			t.mu.Lock()
			s.done <- struct{}{}
			t.mu.Unlock()
		}
	}
	return nil
}
