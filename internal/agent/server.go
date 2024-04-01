package agent

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type commandServer struct {
	UnimplementedCommandExecutorServer
}

func (s *commandServer) Execute(ctx context.Context, r *ExecuteCommandRequest) (*ExecuteCommandResponse, error) {
	res, err := s.executeCommand(ctx, Command(r.Command), "TODO: args")
	if err != nil {
		return nil, err
	}

	return &ExecuteCommandResponse{Result: res.Payload}, nil
}

func ListenAndServe(port string) error {

	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	cmdServer := new(commandServer)
	RegisterCommandExecutorServer(server, cmdServer)

	return server.Serve(l)
}

// https://github.com/shirou/gopsutil

type Command string

func (c Command) String() string {
	return string(c)
}

const (
	CommandCheckCPU        Command = "check_cpu"
	CommandCheckRAM        Command = "check_ram"
	CommandCheckDisk       Command = "check_disk"
	CommandCheckRessources Command = "check_ressources" // checks cpu, ram and disk in one call
	CommandCheckDocker     Command = "check_docker"
)

type Result struct {
	Payload []byte `json:"payload"`
}

func (s *commandServer) executeCommand(ctx context.Context, cmd Command, args string) (*Result, error) {

	var executor Executor
	var ok bool
	if executor, ok = executors[cmd]; !ok {
		return nil, fmt.Errorf("invalid command '%s'", cmd)
	}

	return executor.Execute(ctx)
}
