package cri

import (
	"context"
	"io"
	"log"
	"os/exec"
)

type Client interface {
	Exec(ctx context.Context, containerID string, command []string, stdin io.Reader) (string, error)
}

type CliClient struct {
	endpoint string
}

type Endpoint string

func NewCliClient(endpoint Endpoint) *CliClient {
	return &CliClient{endpoint: string(endpoint)}
}

func (c *CliClient) Exec(ctx context.Context, containerID string, command []string, stdin io.Reader) (string, error) {
	args := []string{"-r", c.endpoint, "exec", "-i", containerID}
	args = append(args, command...)
	log.Printf("huh %q", c.endpoint)
	cmd := exec.CommandContext(ctx, "crictl", args...)
	cmd.Stdin = stdin
	output, err := cmd.CombinedOutput()
	return string(output), err
}
