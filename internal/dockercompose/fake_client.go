package dockercompose

import (
	"context"
	"io"
)

type FakeDCClient struct {
	UpCounts map[string]int
	UpStdout string
	UpStderr string
}

// TODO(dmiller) make this configurable for testing
func NewFakeDockerComposeClient() DockerComposeClient {
	return &FakeDCClient{}
}

func (c *FakeDCClient) Up(ctx context.Context, pathToConfig, serviceName string, stdout, stderr io.Writer) error {
	return nil
}

func (c *FakeDCClient) Down(ctx context.Context, pathToConfig string, stdout, stderr io.Writer) error {
	return nil
}

func (c *FakeDCClient) Logs(ctx context.Context, pathToConfig, serviceName string) (io.ReadCloser, error) {
	return nil, nil
}

func (c *FakeDCClient) Events(ctx context.Context, pathToConfig string) (<-chan string, error) {
	return nil, nil
}

func (c *FakeDCClient) Config(ctx context.Context, pathToConfig string) (string, error) {
	return "", nil
}

func (c *FakeDCClient) Services(ctx context.Context, pathToConfig string) (string, error) {
	return "", nil
}
