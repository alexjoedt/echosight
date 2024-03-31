package echosight

import (
	"context"

	"github.com/alexjoedt/echosight/internal/eventflow"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/google/uuid"
)

// DetectorService
type DetectorService interface {
	Create(ctx context.Context, detector *Detector) error
	GetByID(ctx context.Context, id uuid.UUID) (*Detector, error)
	GetByName(ctx context.Context, name string) (*Detector, error)
	Update(ctx context.Context, detector *Detector) error
	DeleteByID(ctx context.Context, id uuid.UUID) (*Detector, error)
	List(ctx context.Context, detectorFilter *filter.DetectorFilter) ([]*Detector, error)
}

// HostService
type HostService interface {
	Create(ctx context.Context, host *Host) error
	GetByID(ctx context.Context, id uuid.UUID) (*Host, error)
	GetByName(ctx context.Context, name string) (*Host, error)
	Update(ctx context.Context, host *Host) error
	DeleteByID(ctx context.Context, id uuid.UUID) (*Host, error)
	List(ctx context.Context, hostFilter *filter.HostFilter) ([]*Host, error)
}

// UserService represents a service for managing users.
type UserService interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	DeleteByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, userFilter *filter.UserFilter) ([]*User, error)
}

// RecipientService represents a service for managing recipients.
type RecipientService interface {
	Create(ctx context.Context, rcpt *Recipient) error
	GetByID(ctx context.Context, id uuid.UUID) (*Recipient, error)
	GetByEmail(ctx context.Context, email string) (*Recipient, error)
	Update(ctx context.Context, rcpt *Recipient) error
	DeleteByID(ctx context.Context, id uuid.UUID) (*Recipient, error)
	List(ctx context.Context, rcptFilter *filter.RecipientFilter) ([]*Recipient, error)
}

// PreferenceService retrieve and set preferences
// INFO: Preferences are a map and Preference is a struct type
type PreferenceService interface {
	AllPreferences(context.Context) (*Preferences, error)
	GetByName(ctx context.Context, name string) (*Preference, error)
	List(ctx context.Context, prefFilter *filter.PreferenceFilter) (*Preferences, error)
	Set(ctx context.Context, pref *Preference) error
	Update(ctx context.Context, pref *Preference) error
	SetAll(ctx context.Context, prefs *Preferences) error
	DeleteByName(ctx context.Context, name string) error
}

type SessionService interface {
	Put(ctx context.Context, session *Session) error
	Get(ctx context.Context, token string) (*Session, bool, error)
	Delete(ctx context.Context, token string) error
}

type MetricService interface {
	MetricWriter
	MetricReader
}

type MetricWriter interface {
	Write(context.Context, *Metric) error
}

type MetricReader interface {
	Read(context.Context, *MetricFilter) ([]MetricPoint, error)
}

type EventHandler interface {
	EventPusher
	EventSubscriber
}

type EventPusher interface {
	Push(ctx context.Context, topicID string, event *eventflow.Event) error
}

type EventSubscriber interface {
	Subscribe(ctx context.Context, topicID string, onEventFn EventHandler) (*eventflow.Subscription, error)
}

type Mailer interface {
	SendTemplate(templateFile string, data any) error
}

type Crypter interface {
	Encrypt(text string) (string, error)
	Decrypt(ciphertextHex string) (string, error)
}
