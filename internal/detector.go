package echosight

import (
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/uptrace/bun"
)

// Detector represents a Detector to check several endpoints
// A Detector holds common information and is the database model.
type Detector struct {
	bun.BaseModel `bun:"table:detectors"`
	ID            uuid.UUID    `json:"id" bun:"type:uuid,pk,default:uuid_generate_v4()"`
	HostID        uuid.UUID    `json:"hostId" bun:"type:uuid"`
	Name          string       `json:"name"`
	HostName      string       `json:"hostName"`
	Active        bool         `json:"active"`
	Type          DetectorType `json:"type"`
	Timeout       Duration     `json:"timeout"`
	Interval      Duration     `json:"interval"`

	Tags []string `json:"tags" bun:",array"`

	// Config is the configuration for the specified checker type which implements also the Checker interface
	Config CheckerConfig `json:"config" bun:"type:jsonb"`

	State         State     `json:"state"`
	StatusMessage string    `json:"statusMessage"`
	LastCheckedAt time.Time `json:"lastCheckedAt"`

	LookupVersion int       `json:"lookupVersion" bun:",default:1"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`

	Metrics []MetricPoint `json:"metrics" bun:"-"`
}

func (d *Detector) ApplyIDs(r *Result) *Result {
	if r == nil {
		return nil
	}

	// TODO: how to pass mesearument and how to group?
	if r.Metric == nil {
		r.Metric = &Metric{
			Tags:   make(map[string]string, 0),
			Fields: make(map[string]any, 0),
			Time:   time.Now(),
		}
	} else if r.Metric != nil && r.Metric.Tags == nil {
		r.Metric.Tags = make(map[string]string, 0)
	} else if r.Metric != nil && r.Metric.Fields == nil {
		// Field must come from the checker
		r.Metric.Fields = make(map[string]any, 0)
	}

	r.Metric.Bucket = "host_metrics"
	r.Metric.Measurement = string(d.Type)
	r.Metric.Tags["host_id"] = d.HostID.String()
	r.Metric.Tags["detector_id"] = d.ID.String()
	return r
}

// TODO: should we split it in different buckets?
func (d *Detector) MetricFiler(timeRange string) *MetricFilter {
	return &MetricFilter{
		DetectorID: d.ID.String(),
		HostID:     d.HostID.String(),
		Bucket:     "host_metrics",
		Range:      timeRange,
		Mesurement: string(d.Type),
	}
}

func ValidateDetectorConfig(d *Detector) bool {
	if d.Config == nil {
		return false
	}

	// Add here new Checkers/Detectors
	// TODO:  improve this
	switch d.Type {
	case DetectorHTTP:
		var httpConfig HTTPChecker
		return nil == mapstructure.Decode(d.Config, &httpConfig)
	case DetectorPostgres:
		var postgresConfig PostgresChecker
		return nil == mapstructure.Decode(d.Config, &postgresConfig)
	case DetectorAgent:
		var agentConfig AgentConfig
		return nil == d.Config.Unmarshal(&agentConfig)
	}

	return false
}
