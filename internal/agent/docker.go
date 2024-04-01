package agent

import (
	"context"
	"encoding/json"
)

type DockerResult struct {
	Total       uint64
	Free        uint64
	Used        uint64
	UsedPercent float64
}

func (m *DockerResult) Bytes() []byte {
	data, _ := json.MarshalIndent(m, "", "\t")
	return data
}

var _ Executor = (*DockerExecutor)(nil)

type DockerExecutor struct{}

func (e *DockerExecutor) Execute(ctx context.Context) (*Result, error) {
	// TODO: implement
	return &Result{}, nil
}
