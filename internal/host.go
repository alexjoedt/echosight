package echosight

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AddressType string

func (at AddressType) String() string {
	return string(at)
}

const (
	AddressTypeIPv4 AddressType = "IPv4"
	AddressTypeIPv6 AddressType = "IPv6"
)

// A host can be observed with an agent, who runs as a daemon on the host
// the agent (see commander package) responses with metrics such as cpu, ram, disk
// The metrics are saved at influxdb

// Host
type Host struct {
	bun.BaseModel `bun:"table:hosts"`
	ID            uuid.UUID   `json:"id" bun:"type:uuid,pk,default:uuid_generate_v4()"`
	LookupVersion int         `json:"lookupVersion" bun:",default:1"`
	Name          string      `json:"name"`
	AddressType   AddressType `json:"addressType"` // IPv4 oder IPv6
	Address       string      `json:"address"`
	Location      string      `json:"location"`
	OS            string      `json:"os"`
	Active        bool        `json:"active"`
	Agent         bool        `json:"agent"`
	State         State       `json:"state"`
	StatusMessage string      `json:"statusMessage"`
	LastCheckedAt time.Time   `json:"lastCheckedAt"`
	Tags          []string    `json:"tags" bun:",array"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Detectors []*Detector `json:"detectors,omitempty" bun:"rel:has-many,join:id=host_id"`
}
