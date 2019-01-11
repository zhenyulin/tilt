package engine

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/windmilleng/tilt/internal/container"
	"github.com/windmilleng/tilt/internal/docker"
	"github.com/windmilleng/tilt/internal/logger"
	"github.com/windmilleng/tilt/internal/model"
	"github.com/windmilleng/tilt/internal/store"
)

// Collects logs from running docker-compose services.
type DockerComposeLogManager struct {
	watches map[model.ManifestName]dockerComposeLogWatch
	dCli    docker.DockerClient
}

func NewDockerComposeLogManager(dCli docker.DockerClient) *DockerComposeLogManager {
	return &DockerComposeLogManager{
		watches: make(map[model.ManifestName]dockerComposeLogWatch),
		dCli:    dCli,
	}
}

// Diff the current watches against set of current docker-compose services, i.e.
// what we SHOULD be watching, returning the changes we need to make.
func (m *DockerComposeLogManager) diff(ctx context.Context, st store.RStore) (setup []dockerComposeLogWatch, teardown []dockerComposeLogWatch) {
	state := st.RLockState()
	defer st.RUnlockState()

	// If we're not watching the mounts, then don't bother watching logs.
	if !state.WatchMounts {
		return nil, nil
	}

	for _, mt := range state.ManifestTargets {
		manifest := mt.Manifest
		if !manifest.IsDC() {
			continue
		}

		cID := mt.State.DCResourceState().ContainerID

		existing, isActive := m.watches[manifest.Name]
		startWatchTime := time.Unix(0, 0)
		if isActive {
			select {
			case termTime := <-existing.terminationTime:
				// If we're receiving on this channel, it's because the previous watcher ended or
				// died somehow; we need to create a new one that picks up where it left off.
				startWatchTime = termTime
			default:
				// Watcher is still active, no action needed.
				continue
			}
		}

		ctx, cancel := context.WithCancel(ctx)
		w := dockerComposeLogWatch{
			ctx:             ctx,
			cancel:          cancel,
			name:            manifest.Name,
			cID:             cID,
			startWatchTime:  startWatchTime,
			terminationTime: make(chan time.Time, 1),
		}
		m.watches[manifest.Name] = w
		setup = append(setup, w)
	}

	for key, value := range m.watches {
		_, inState := state.ManifestTargets[key]
		if !inState {
			delete(m.watches, key)

			teardown = append(teardown, value)
		}
	}

	return setup, teardown
}

func (m *DockerComposeLogManager) OnChange(ctx context.Context, st store.RStore) {
	setup, teardown := m.diff(ctx, st)
	for _, watch := range teardown {
		watch.cancel()
	}

	for _, watch := range setup {
		go m.consumeLogs(watch, st)
	}
}

func (m *DockerComposeLogManager) consumeLogs(watch dockerComposeLogWatch, st store.RStore) {
	defer func() {
		watch.terminationTime <- time.Now()
	}()

	name := watch.name
	readCloser, err := m.dCli.StreamLogs(watch.ctx, watch.cID, watch.startWatchTime)
	if err != nil {
		logger.Get(watch.ctx).Debugf("Error streaming %s logs: %v", name, err)
		return
	}
	defer func() {
		_ = readCloser.Close()
	}()

	// TODO(maia): docker-compose already prefixes logs, but maybe we want to roll
	// our own (as in PodWatchManager) cuz it's prettier?
	globalLogWriter := DockerComposeGlobalLogWriter{
		writer: logger.Get(watch.ctx).Writer(logger.InfoLvl),
	}
	actionWriter := DockerComposeLogActionWriter{
		store:        st,
		manifestName: name,
	}
	multiWriter := io.MultiWriter(globalLogWriter, actionWriter)

	_, err = io.Copy(multiWriter, NewHardCancelReader(watch.ctx, readCloser))
	if err != nil && watch.ctx.Err() == nil {
		logger.Get(watch.ctx).Debugf("Error streaming %s logs: %v", name, err)
		return
	}
}

type dockerComposeLogWatch struct {
	ctx             context.Context
	cancel          func()
	name            model.ManifestName
	cID             container.ID
	startWatchTime  time.Time
	terminationTime chan time.Time
}

type DockerComposeLogActionWriter struct {
	store        store.RStore
	manifestName model.ManifestName
}

func (w DockerComposeLogActionWriter) Write(p []byte) (n int, err error) {
	if shouldFilterDCLog(p) {
		return len(p), nil
	}
	w.store.Dispatch(DockerComposeLogAction{
		ManifestName: w.manifestName,
		Log:          append([]byte{}, p...),
	})
	return len(p), nil
}

var _ store.Subscriber = &DockerComposeLogManager{}

func shouldFilterDCLog(p []byte) bool {
	if bytes.HasPrefix(p, []byte("Attaching to ")) {
		return true
	}

	return false
}

type DockerComposeGlobalLogWriter struct {
	writer io.Writer
}

func (w DockerComposeGlobalLogWriter) Write(p []byte) (n int, err error) {
	if shouldFilterDCLog(p) {
		return len(p), nil
	}

	return w.writer.Write(p)
}
