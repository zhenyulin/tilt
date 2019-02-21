// +build wireinject
// The build tag makes sure the stub is not built in the final build.

package synclet

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/wire"
	"google.golang.org/grpc"
	criproto "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"

	"github.com/windmilleng/tilt/internal/container"
	"github.com/windmilleng/tilt/internal/docker"
	"github.com/windmilleng/tilt/internal/k8s"
	"github.com/windmilleng/tilt/internal/k8s/cri"
)

func WireDockerSynclet(ctx context.Context, env k8s.Env, runtime container.Runtime) (*DockerSynclet, error) {
	wire.Build(
		docker.DefaultClient,
		wire.Bind(new(docker.Client), new(docker.Cli)),

		NewDockerSynclet,
	)
	return nil, nil
}

func WireCriSynclet(ctx context.Context, endpoint Endpoint) (*CriSynclet, error) {
	wire.Build(
		provideRuntimeClient,
		cri.NewGrpcClient,
		wire.Bind(new(cri.Client), new(cri.GrpcClient)),
		NewCriSynclet,
	)
	return nil, nil
}

type Endpoint string

func provideRuntimeClient(ctx context.Context, endpoint Endpoint) (criproto.RuntimeServiceClient, error) {

	// hrm
	addr := string(endpoint)

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second), grpc.WithDialer(dial))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %v: %v", addr, err)
	}

	return criproto.NewRuntimeServiceClient(conn), nil
}

func dial(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", addr, timeout)
}
