package agent

import (
	"context"
	"encoding/json"

	"github.com/shirou/gopsutil/v3/mem"
)

type MemoryResult struct {
	Total       uint64
	Free        uint64
	Used        uint64
	UsedPercent float64
}

func (m *MemoryResult) Bytes() []byte {
	data, _ := json.MarshalIndent(m, "", "\t")
	return data
}

var _ Executor = (*MemoryExecutor)(nil)

type MemoryExecutor struct{}

func (e *MemoryExecutor) Execute(ctx context.Context) (*Result, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	res := &MemoryResult{
		Total:       v.Total,
		Free:        v.Free,
		Used:        v.Used,
		UsedPercent: v.UsedPercent,
	}

	return &Result{Payload: res.Bytes()}, nil
}
