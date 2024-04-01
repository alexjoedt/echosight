package echosight

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
)

type Checker interface {
	ID() string
	Check(ctx context.Context) *Result
	Interval() time.Duration
	Detector() *Detector
}

type CheckerConfig map[string]any

func (dc *CheckerConfig) Unmarshal(v any) error {
	return mapstructure.Decode(dc, v)
}

// TODO: use int with iota, see below?
type State string

const (
	StateOK       State = "OK"
	StateWarn     State = "WARN"
	StateCritical State = "CRITICAL"
	StateInactive State = "INACTIVE"
)

// TODO: use this?
type StateInt int

const (
	StateIntInactive StateInt = iota - 1
	StateIntOK
	StateIntWarn
	StateIntCritical
)

func (s State) String() string {
	return string(s)
}

var (
	checkerFuncs = map[DetectorType]func(d *Detector) (Checker, error){
		DetectorHTTP:     getHTTPChecker,
		DetectorAgent:    getAgentChecker,
		DetectorPostgres: getPostgresChecker,
	}
)

func getHTTPChecker(d *Detector) (Checker, error) {
	var httpDetector HTTPChecker
	err := d.Config.Unmarshal(&httpDetector)
	if err != nil {
		return nil, err
	}

	httpDetector.detector = d
	return &httpDetector, nil
}

func getAgentChecker(d *Detector) (Checker, error) {
	var agentChecker AgentConfig
	err := d.Config.Unmarshal(&agentChecker)
	if err != nil {
		return nil, err
	}

	if agentChecker.WarnThreshold == 0 {
		agentChecker.WarnThreshold = 85
	}

	if agentChecker.CriticalThreshold == 0 {
		agentChecker.CriticalThreshold = 95
	}

	agentChecker.detector = d
	return &agentChecker, nil
}

func getPostgresChecker(d *Detector) (Checker, error) {
	var postgresDetector PostgresChecker
	err := d.Config.Unmarshal(&postgresDetector)
	if err != nil {
		return nil, err
	}

	postgresDetector.detector = d
	return &postgresDetector, nil
}

func (d *Detector) GetChecker() (Checker, error) {
	fn, ok := checkerFuncs[d.Type]
	if !ok {
		return nil, fmt.Errorf("invalid detector type: '%s'", d.Type)
	}

	return fn(d)
}

type Result struct {
	Host     string  `json:"host"`
	Detector string  `json:"detector"`
	State    State   `json:"state"`
	Message  string  `json:"message"`
	Metric   *Metric `json:"metric"`

	// err indicates if an internal err happened
	err error
}

func (r *Result) Error() error {
	return r.err
}

func (r *Result) String() string {
	return fmt.Sprintf("state=%s, message=%v, metric=%v", r.State, r.Message, r.Metric)
}
