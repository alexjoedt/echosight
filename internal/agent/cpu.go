package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

type CPUResult struct {
	CPUs map[string]float64
}

func (m *CPUResult) Bytes() []byte {
	data, _ := json.MarshalIndent(m, "", "\t")
	return data
}

type CPUExecutor struct {
	allCPUs bool
}

func NewCPUExecutor(args string) *CPUExecutor {
	executor := new(CPUExecutor)

	return executor
}

func (e *CPUExecutor) Execute(ctx context.Context) (*Result, error) {
	c, _ := cpu.Percent(time.Millisecond*500, true)
	res := CPUResult{
		CPUs: make(map[string]float64, len(c)),
	}

	for i, r := range c {
		res.CPUs[fmt.Sprintf("cpu_%d", i)] = r
	}

	return &Result{Payload: res.Bytes()}, nil
}
