package echosight

import (
	"fmt"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Metric struct {
	Bucket      string
	Measurement string
	Fields      map[string]any
	Tags        map[string]string
	Time        time.Time
}

func (m *Metric) Point() *write.Point {
	return write.NewPoint(m.Measurement, m.Tags, m.Fields, m.Time)
}

type MetricFilter struct {
	DetectorID string
	HostID     string
	Bucket     string
	Range      string
	Mesurement string
	Field      string
	Yield      string
}

func (mf *MetricFilter) Query() string {
	// 	query := `from(bucket: "host_metrics")
	// |> range(start: -60m)
	// |> filter(fn: (r) => r["_measurement"] == "http")
	// |> yield(name: "mean")`
	fieldQuery := ""
	if mf.HostID != "" && mf.DetectorID != "" {
		fieldQuery = fmt.Sprintf(` and r.detector_id == "%s" and r.host_id == "%s"`, mf.DetectorID, mf.HostID)
	}

	raw := `from(bucket: "%s")
	|> range(start: %s)
	|> filter(fn: (r) => r._measurement == "%s"%s)
	|> yield(name: "%s")`

	return fmt.Sprintf(raw, mf.Bucket, mf.Range, mf.Mesurement, fieldQuery, mf.Yield)
}

type MetricPoint struct {
	Time   time.Time      `json:"time"`
	Fields map[string]any `json:"fields"`
}
