package engine

import (
	"github.com/windmilleng/tilt/internal/model"
)

type internalState struct {
	k8s map[string]*k8sResource
}

type k8sResource struct {
	manifest model.Manifest
	bs       BuildState
}
