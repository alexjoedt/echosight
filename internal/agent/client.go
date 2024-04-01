package agent

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	c    CommandExecutorClient
	conn *grpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := NewCommandExecutorClient(conn)

	return &Client{
		c:    c,
		conn: conn,
	}, nil
}

func (c *Client) execute(ctx context.Context, cmd Command) (*Result, error) {
	response, err := c.c.Execute(ctx, &ExecuteCommandRequest{
		Command: cmd.String(),
	})
	if err != nil {
		return nil, err
	}

	return &Result{Payload: response.GetResult()}, nil
}

func (c *Client) CheckCPU(ctx context.Context) (*CPUResult, error) {
	res, err := c.execute(ctx, CommandCheckCPU)
	if err != nil {
		return nil, err
	}

	var cpuResult CPUResult
	err = json.Unmarshal(res.Payload, &cpuResult)
	if err != nil {
		return nil, err
	}

	return &cpuResult, nil
}

func (c *Client) CheckMemory(ctx context.Context) (*MemoryResult, error) {
	res, err := c.execute(ctx, CommandCheckRAM)
	if err != nil {
		return nil, err
	}

	var ramResult MemoryResult
	err = json.Unmarshal(res.Payload, &ramResult)
	if err != nil {
		return nil, err
	}

	return &ramResult, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
