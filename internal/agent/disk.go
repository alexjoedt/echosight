package agent

import (
	"context"
	"encoding/json"
)

type DiskResult struct {
	Total       uint64
	Free        uint64
	Used        uint64
	UsedPercent float64
}

func (m *DiskResult) Bytes() []byte {
	data, _ := json.MarshalIndent(m, "", "\t")
	return data
}

var _ Executor = (*DiskExecutor)(nil)

type DiskExecutor struct{}

func (e *DiskExecutor) Execute(ctx context.Context) (*Result, error) {
	// TODO: implement

	// res := &DiskResult{
	// 	Total:       v.Total,
	// 	Free:        v.Free,
	// 	Used:        v.Used,
	// 	UsedPercent: v.UsedPercent,
	// }

	return &Result{}, nil
}
