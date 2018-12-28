package model

import (
	"sort"

	"github.com/docker/distribution/reference"
)

// BuildInfo holds the info for how to build images
// Right now we only build with docker in some way,
// But we also allow not building, in which case it will be nil.
type BuildInfo interface {
	buildInfo()
}

type DockerInfo struct {
	cachePaths []string
	DockerRef  reference.Named
	Details    buildDetails
}

type buildDetails interface {
	buildDetails()
}

func (di DockerInfo) WithCachePaths(paths []string) DockerInfo {
	di.cachePaths = append(append([]string{}, di.cachePaths...), paths...)
	sort.Strings(di.cachePaths)
	return di
}

func (di DockerInfo) CachePaths() []string {
	return di.cachePaths
}

func (DockerInfo) buildInfo() {}

func (di DockerInfo) StaticBuild() *StaticBuild {
	switch bd := di.Details.(type) {
	case StaticBuild:
		return &bd
	default:
		return nil
	}
}

func (di DockerInfo) FastBuild() *FastBuild {
	switch bd := di.Details.(type) {
	case FastBuild:
		return &bd
	default:
		return nil
	}
}

type StaticBuild struct {
	Dockerfile string
	BuildPath  string // the absolute path to the files
	BuildArgs  DockerBuildArgs
}

func (StaticBuild) buildDetails() {}

type FastBuild struct {
	BaseDockerfile string
	Mounts         []Mount
	Steps          []Step
	Entrypoint     Cmd
}

func (FastBuild) buildDetails() {}
