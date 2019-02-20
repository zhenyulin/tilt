package sidecar

import (
	"fmt"

	"github.com/docker/distribution/reference"

	"github.com/windmilleng/tilt/internal/container"
	"github.com/windmilleng/tilt/internal/k8s"
)

func InjectSyncletSidecar(entity k8s.K8sEntity, matchRef reference.Named, runtime container.Runtime) (k8s.K8sEntity, bool, error) {
	entity = entity.DeepCopy()

	pods, err := k8s.ExtractPods(&entity)
	if err != nil {
		return k8s.K8sEntity{}, false, err
	}

	replaced := false
	for _, pod := range pods {
		ok, err := k8s.PodContainsRef(*pod, matchRef)
		if err != nil {
			return k8s.K8sEntity{}, false, err
		}

		if !ok {
			continue
		}

		replaced = true

		containerAndVol, ok := RuntimeToConfig[runtime]
		if !ok {
			return k8s.K8sEntity{}, false, fmt.Errorf("Don't have a synclet for container runtime %q", runtime)
		}

		vol := containerAndVol.Volume.DeepCopy()
		pod.Volumes = append(pod.Volumes, *vol)

		container := containerAndVol.Container.DeepCopy()
		pod.Containers = append(pod.Containers, *container)
	}
	return entity, replaced, nil
}
