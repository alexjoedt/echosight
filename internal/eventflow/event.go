package eventflow

import (
	"context"
	"encoding/json"
)

type EventHandler func(ctx context.Context, event *Event) error

type Event struct {
	Type           string          `json:"eventType"`
	Payload        json.RawMessage `json:"payload"`
	topicID        string
	subscriptionID string
}

func (e *Event) TopicID() string {
	return e.topicID
}

func (e *Event) SubscriptionID() string {
	return e.subscriptionID
}
