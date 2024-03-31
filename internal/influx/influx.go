package influx

import (
	"context"
	"sort"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	influx "github.com/influxdata/influxdb-client-go/v2"
)

var _ echosight.MetricService = (*InfluxDB)(nil)

const (
	org string = "echosight"
)

type InfluxDB struct {
	client influx.Client
}

// TODO: setup influx at first start and safe the state in the postgres database
// state should be the authtoken and that influx is setup

// New creates a new influx db client
func New(url string, token string) (*InfluxDB, error) {
	client := influx.NewClientWithOptions(url, token, influx.DefaultOptions().SetBatchSize(20))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := client.Health(ctx)
	if err != nil {
		return nil, err
	}

	return &InfluxDB{
		client: client,
	}, nil
}

func (i *InfluxDB) Write(ctx context.Context, metric *echosight.Metric) error {
	return i.client.WriteAPIBlocking(org, metric.Bucket).WritePoint(ctx, metric.Point())
}

func (i *InfluxDB) Read(ctx context.Context, mFilter *echosight.MetricFilter) ([]echosight.MetricPoint, error) {
	result, err := i.client.QueryAPI(org).Query(ctx, mFilter.Query())
	if err != nil {
		return nil, err
	}

	resultPoints := make(map[time.Time]echosight.MetricPoint, 0)

	for result.Next() {
		val, ok := resultPoints[result.Record().Time()]
		if !ok {
			val = echosight.MetricPoint{
				Time:   result.Record().Time(),
				Fields: make(map[string]any, 0),
			}
		}

		val.Fields[result.Record().Field()] = result.Record().Value()
		resultPoints[result.Record().Time()] = val
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return sortPointsByTime(resultPoints), nil
}

func sortPointsByTime(resultPoints map[time.Time]echosight.MetricPoint) []echosight.MetricPoint {
	var keys []time.Time
	for k := range resultPoints {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(keys[j])
	})

	var sortedPoints []echosight.MetricPoint
	for _, k := range keys {
		sortedPoints = append(sortedPoints, resultPoints[k])
	}

	return sortedPoints
}
