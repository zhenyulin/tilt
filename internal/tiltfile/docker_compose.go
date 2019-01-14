package tiltfile

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/stack/loader"
	"github.com/docker/cli/cli/command/stack/options"
	"github.com/docker/cli/cli/compose/types"
	"github.com/pkg/errors"
	"github.com/windmilleng/tilt/internal/dockerfile"
	"github.com/windmilleng/tilt/internal/model"
	"go.starlark.net/starlark"
	"gopkg.in/yaml.v2"
)

// dcResource represents a single docker-compose config file and all its associated services
type dcResource struct {
	configPath string

	services []dcService
}

func (dc dcResource) Empty() bool { return reflect.DeepEqual(dc, dcResource{}) }

func (s *tiltfileState) dockerCompose(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var configPath string
	err := starlark.UnpackArgs(fn.Name(), args, kwargs, "configPath", &configPath)
	if err != nil {
		return nil, err
	}
	configPath = s.absPath(configPath)
	if err != nil {
		return nil, err
	}

	services, err := parseDCConfig(s.ctx, configPath)
	if err != nil {
		return nil, err
	}

	if !s.dc.Empty() {
		return starlark.None, fmt.Errorf("already have a docker-compose resource declared (%s), cannot declare another (%s)", s.dc.configPath, configPath)
	}

	s.dc = dcResource{configPath: configPath, services: services}

	return starlark.None, nil
}

// A docker-compose service, according to Tilt.
type dcService struct {
	Name             string
	Context          string
	DfPath           string
	MountedLocalDirs []string

	// Currently just use these to diff against when config files are edited to see if manifest has changed
	ServiceConfig []byte
	DfContents    []byte
}

func dockerServiceToModel(config types.ServiceConfig) (dcService, error) {
	df := config.Build.Dockerfile
	if df == "" && config.Build.Context != "" {
		// We only expect a Dockerfile if there's a build context specified.
		df = "Dockerfile"
	}

	var mountedLocalDirs []string
	for _, v := range config.Volumes {
		mountedLocalDirs = append(mountedLocalDirs, v.Source)
	}

	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		return dcService{}, errors.Wrapf(err, "error serializing docker-compose config for %s", config.Name)
	}

	dfPath := filepath.Join(config.Build.Context, df)
	svc := dcService{
		Name:             config.Name,
		Context:          config.Build.Context,
		DfPath:           dfPath,
		MountedLocalDirs: mountedLocalDirs,
		ServiceConfig:    yamlBytes,
	}

	if dfPath != "" {
		dfContents, err := ioutil.ReadFile(dfPath)
		if err != nil {
			return svc, err
		}
		svc.DfContents = dfContents
	}
	return svc, nil
}

func parseDCConfig(ctx context.Context, configPath string) ([]dcService, error) {
	w := &bytes.Buffer{}
	cli := command.NewDockerCli(nil, w, w, true, nil)
	dcOpts := options.Deploy{
		Composefiles: []string{configPath},
	}
	configDetails, err := loader.LoadComposefile(cli, dcOpts)
	if err != nil {
		return nil, errors.Wrap(err, "loading compose file")
	}

	var services []dcService

	for _, dockerSvc := range configDetails.Services {
		svc, err := dockerServiceToModel(dockerSvc)
		if err != nil {
			return nil, errors.Wrapf(err, "getting service %s", dockerSvc.Name)
		}
		services = append(services, svc)
	}

	return services, nil
}

func (s *tiltfileState) dcServiceToManifest(service dcService, dcConfigPath string) (manifest model.Manifest,
	configFiles []string, err error) {
	dcInfo := model.DockerComposeTarget{
		ConfigPath: dcConfigPath,
		DfRaw:      service.DfContents,
		YAMLRaw:    service.ServiceConfig,
	}

	m := model.Manifest{
		Name: model.ManifestName(service.Name),
	}.WithDeployTarget(dcInfo)

	if service.DfPath == "" {
		// DC service may not have Dockerfile -- e.g. may be just an image that we pull and run.
		// So, don't parse a non-existent Dockerfile for mount info.
		return m, nil, nil
	}

	df := dockerfile.Dockerfile(service.DfContents)
	mounts, err := df.DeriveMounts(service.Context)
	if err != nil {
		return model.Manifest{}, nil, err
	}

	dcInfo.Mounts = mounts

	paths := []string{path.Dir(service.DfPath), path.Dir(dcConfigPath)}
	for _, mount := range mounts {
		paths = append(paths, mount.LocalPath)
	}

	dcInfo = dcInfo.WithDockerignores(dockerignoresForPaths(append(paths, path.Dir(s.filename.path))))

	localPaths := []localPath{s.filename}
	for _, p := range paths {
		localPaths = append(localPaths, s.localPathFromString(p))
	}
	dcInfo = dcInfo.WithRepos(reposForPaths(localPaths)).
		WithTiltFilename(s.filename.path).
		WithIgnoredLocalDirectories(service.MountedLocalDirs)

	m = m.WithDeployTarget(dcInfo).
		WithTiltFilename(s.filename.path)

	return m, []string{service.DfPath}, nil
}
