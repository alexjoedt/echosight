package echosight

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alexjoedt/echosight/internal/agent"
)

var (
	_ Checker   = (*AgentConfig)(nil)
	_ Validator = (*AgentConfig)(nil)
)

const (
	defaultPort string = "8089"
)

type AgentConfig struct {
	IP      string        `json:"ip"`
	Command agent.Command `json:"command"`
	Count   int           `json:"count,omitempty"`

	WarnThreshold     float64 `json:"warnThreshold"`
	CriticalThreshold float64 `json:"criticalThreshold"`

	client   *agent.Client
	detector *Detector `json:"-"`
}

func (a *AgentConfig) Validate() bool {
	// add valid commands here
	switch a.Command {
	case agent.CommandCheckCPU:
	case agent.CommandCheckDisk:
	case agent.CommandCheckRAM:
	case agent.CommandCheckRessources:
	case agent.CommandCheckDocker:
	default:
		return false
	}

	return nil != net.ParseIP(a.IP)
}

func (a *AgentConfig) ID() string {
	return a.detector.ID.String()
}

func (a *AgentConfig) Interval() time.Duration {
	return time.Duration(a.detector.Interval)
}

func (a *AgentConfig) Detector() *Detector {
	return a.detector
}

func (a *AgentConfig) Check(ctx context.Context) *Result {
	address := a.IP
	if !strings.Contains(address, ":") {
		address = fmt.Sprintf("%s:%s", a.IP, defaultPort)
	}

	var err error
	a.client, err = agent.NewClient(address)
	if err != nil {
		return &Result{State: StateCritical, Message: err.Error(), err: err}
	}
	defer a.client.Close()

	var result *Result
	// TODO: create specific funcs and assign to a map, like the checker funcs
	switch a.Command {
	case agent.CommandCheckCPU:
		result = a.checkCPU(ctx)
	case agent.CommandCheckRAM:
		return &Result{State: StateCritical, Message: "not implemented"}
	case agent.CommandCheckDisk:
		return &Result{State: StateCritical, Message: "not implemented"}
	case agent.CommandCheckDocker:
		return &Result{State: StateCritical, Message: "not implemented"}
	case agent.CommandCheckRessources:
		return &Result{State: StateCritical, Message: "not implemented"}
	default:
		return &Result{State: StateCritical, Message: "not implemented"}
	}

	return a.detector.ApplyIDs(result)
}

func (a *AgentConfig) checkCPU(ctx context.Context) *Result {
	res, err := a.client.CheckCPU(ctx)
	if err != nil {
		return &Result{State: StateCritical, Message: err.Error(), err: err}
	}

	result := &Result{
		State: StateOK,
		Metric: &Metric{
			Fields: make(map[string]any, 0),
			Tags:   make(map[string]string, 0),
			Time:   time.Now(),
		},
	}

	for k, v := range res.CPUs {
		result.Metric.Fields[k] = v
	}

	result.State = a.evaluateThreshold(a.calcCPUAverage(result))
	return result
}

// calcCPUAverage calculates the average from multiple cpu cores
func (a *AgentConfig) calcCPUAverage(r *Result) float64 {
	var count float64
	var sum float64

	for _, v := range r.Metric.Fields {
		if i, ok := v.(float64); ok {
			sum += i
			count++
		}
	}

	return sum / count
}

func (a *AgentConfig) evaluateThreshold(value float64) State {
	switch {
	case value >= a.WarnThreshold && value < a.CriticalThreshold:
		return StateWarn
	case value >= a.CriticalThreshold:
		return StateCritical
	default:
		return StateOK
	}
}
