package engine

import (
	"github.com/windmilleng/tilt/internal/model"
	"github.com/windmilleng/tilt/internal/state"
)

type internalState struct {
	pipelineCh chan brAndErr

	k8s             map[string]*k8sResource
	runQueue        []string // resource names
	runningSpan     state.SingleSpanWriter
	runningResource string
	lastSpan        state.SingleSpanWriter
}

type k8sResource struct {
	manifest model.Manifest
	bs       BuildState
}

type brAndErr struct {
	br  BuildResult
	err error
}
