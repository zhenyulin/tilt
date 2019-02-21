package cri

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	criproto "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"

	"github.com/windmilleng/tilt/internal/logger"
)

type Client interface {
	Exec(ctx context.Context, containerID string, command []string, stdin io.Reader) (string, error)
}

type GrpcClient struct {
	cli criproto.RuntimeServiceClient
}

func NewGrpcClient(cli criproto.RuntimeServiceClient) *GrpcClient {
	return &GrpcClient{cli: cli}
}

func (c *GrpcClient) Exec(ctx context.Context, containerId string, command []string, stdin io.Reader) (string, error) {
	req := &criproto.ExecRequest{
		ContainerId: containerId,
		Cmd:         command,
		Stdin:       stdin != nil,
		Stdout:      true,
		Stderr:      true,
	}

	logger.Get(ctx).Infof("about to exec")
	resp, err := c.cli.Exec(ctx, req)
	if err != nil {
		logger.Get(ctx).Infof("error exec'ing: %v", err)
		return "", err
	}

	logger.Get(ctx).Infof("Exec'ed! %+v", resp)

	execUrl, err := url.Parse(resp.Url)
	hostIp := os.Getenv("SYNCLET_HOST_IP")
	if hostIp != "" {
		execUrl.Host = fmt.Sprintf("%s:%s", hostIp, execUrl.Port())
	}

	logger.Get(ctx).Infof("url: %v", execUrl)

	w := logger.Get(ctx).Writer(logger.InfoLvl)

	executor, err := remotecommand.NewSPDYExecutor(&rest.Config{TLSClientConfig: rest.TLSClientConfig{Insecure: true}}, "POST", execUrl)
	if err != nil {

	}

	logger.Get(ctx).Infof("have executor %v", executor)

	opts := remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: w,
		Stderr: w,
	}

	return "", executor.Stream(opts)
}
