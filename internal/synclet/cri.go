package synclet

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"

	"github.com/windmilleng/tilt/internal/container"
	"github.com/windmilleng/tilt/internal/k8s/cri"
	"github.com/windmilleng/tilt/internal/logger"
	"github.com/windmilleng/tilt/internal/model"
)

type CriSynclet struct {
	cri cri.Client
}

func NewCriSynclet(cri cri.Client) *CriSynclet {
	return &CriSynclet{cri: cri}
}

func (s CriSynclet) writeFiles(ctx context.Context, containerId container.ID, tarArchive []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Synclet-writeFiles")
	defer span.Finish()

	if tarArchive == nil {
		return nil
	}

	output, err := s.cri.Exec(ctx, string(containerId), []string{"tar", "-x", "-f", "/dev/stdin"}, bytes.NewBuffer(tarArchive))

	logger.Get(ctx).Infof("writing files; output: %q", output)
	return err
}

func (s CriSynclet) rmFiles(ctx context.Context, containerId container.ID, filesToDelete []string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Synclet-rmFiles")
	defer span.Finish()

	if len(filesToDelete) == 0 {
		return nil
	}

	output, err := s.cri.Exec(ctx, string(containerId), append([]string{"rm"}, filesToDelete...), nil)

	return err
}

func (s CriSynclet) execCmds(ctx context.Context, containerId container.ID, cmds []model.Cmd) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Synclet-execCommands")
	defer span.Finish()

	logger.Get(ctx).Infof("exec'ing %v commands", len(cmds))

	for i, c := range cmds {
		// TODO: instrument this
		logger.Get(ctx).Infof("[CMD %d/%d] %s", i+1, len(cmds), strings.Join(c.Argv, " "))
		// // TODO(matt) - plumb PipelineState through
		err := s.cri.Exec(ctx, string(containerId), c.Argv, nil)
		if err != nil {
			return build.WrapContainerExecError(err, containerId, c)
		}
	}
	return nil
}

func (s CriSynclet) restartContainer(ctx context.Context, containerId container.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Synclet-restartContainer")
	defer span.Finish()

	return fmt.Errorf("CRI doesn't support container restart; you must use Tilt with hot_reload")
}

func (s CriSynclet) UpdateContainer(
	ctx context.Context,
	containerId container.ID,
	tarArchive []byte,
	filesToDelete []string,
	commands []model.Cmd,
	hotReload bool) error {

	span, ctx := opentracing.StartSpanFromContext(ctx, "Synclet-UpdateContainer")
	defer span.Finish()

	err := s.rmFiles(ctx, containerId, filesToDelete)
	if err != nil {
		return fmt.Errorf("error removing files while updating container %s: %v",
			containerId.ShortStr(), err)
	}

	err = s.writeFiles(ctx, containerId, tarArchive)
	if err != nil {
		return fmt.Errorf("error writing files while updating container %s: %v",
			containerId.ShortStr(), err)
	}

	err = s.execCmds(ctx, containerId, commands)
	if err != nil {
		return fmt.Errorf("error exec'ing commands while updating container %s: %v",
			containerId.ShortStr(), err)
	}

	if !hotReload {
		err = s.restartContainer(ctx, containerId)
		if err != nil {
			return fmt.Errorf("error restarting container %s: %v",
				containerId.ShortStr(), err)
		}
	}

	return nil
}
