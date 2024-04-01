package agent

import (
	"context"
)

type Executor interface {
	Execute(ctx context.Context) (*Result, error)
}

var (
	executors = map[Command]Executor{
		CommandCheckCPU:    &CPUExecutor{},
		CommandCheckRAM:    &MemoryExecutor{},
		CommandCheckDisk:   &DiskExecutor{},
		CommandCheckDocker: &DockerExecutor{},
	}
)
