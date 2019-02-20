// +build wireinject
// The build tag makes sure the stub is not built in the final build.

package synclet

import (
	"context"

	"github.com/google/wire"

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

func WireCriSynclet(ctx context.Context, criEndpoint cri.Endpoint) (*CriSynclet, error) {
	wire.Build(
		cri.NewCliClient,
		wire.Bind(new(cri.Client), new(cri.CliClient)),
		NewCriSynclet,
	)
	return nil, nil
}
