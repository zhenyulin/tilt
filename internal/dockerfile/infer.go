package dockerfile

import (
	"fmt"

	"github.com/windmilleng/tilt/internal/model"
)

type Mount struct {
	Src        string
	MountPoint string
}

type FastBuild struct {
	Mounts    []Mount
	Steps     []model.Step
	HotReload bool
}

func Infer(s string) ([]string, *FastBuild, error) {
	return nil, nil, fmt.Errorf("Infer: not yet implemented")
}
