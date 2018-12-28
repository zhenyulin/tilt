package tiltfile2

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/docker/distribution/reference"
	"github.com/google/skylark"

	"github.com/windmilleng/tilt/internal/dockerfile"
	"github.com/windmilleng/tilt/internal/model"
	"github.com/windmilleng/tilt/internal/ospath"
)

type dockerImage interface {
	toDomain() (model.BuildInfo, []model.Dockerignore, []model.LocalGithubRepo)
}

type staticBuild struct {
	dockerInfo model.DockerInfo
	ignores    []model.Dockerignore
	repos      []model.LocalGithubRepo
}

func (s *staticBuild) toDomain() (model.BuildInfo, []model.Dockerignore, []model.LocalGithubRepo) {
	return s.dockerInfo, s.ignores, s.repos
}

func (s *tiltfileState) dockerBuild(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var dockerRef string
	var contextVal, dockerfilePathVal, buildArgs, cacheVal skylark.Value
	if err := skylark.UnpackArgs(fn.Name(), args, kwargs,
		"ref", &dockerRef,
		"context", &contextVal,
		"build_args?", &buildArgs,
		"dockerfile?", &dockerfilePathVal,
		"cache?", &cacheVal,
	); err != nil {
		return nil, err
	}

	ref, err := reference.ParseNormalizedNamed(dockerRef)
	if err != nil {
		return nil, fmt.Errorf("Argument 1 (ref): can't parse %q: %v", dockerRef, err)
	}

	if contextVal == nil {
		return nil, fmt.Errorf("Argument 2 (context): empty but is required")
	}
	context, err := s.localPathFromSkylarkValue(contextVal)
	if err != nil {
		return nil, err
	}

	dockerfilePath := context.join("Dockerfile")
	if dockerfilePathVal != nil {
		dockerfilePath, err = s.localPathFromSkylarkValue(dockerfilePathVal)
		if err != nil {
			return nil, err
		}
	}

	var sba map[string]string
	if buildArgs != nil {
		d, ok := buildArgs.(*skylark.Dict)
		if !ok {
			return nil, fmt.Errorf("Argument 3 (build_args): expected dict, got %T", buildArgs)
		}

		sba, err = skylarkStringDictToGoMap(d)
		if err != nil {
			return nil, fmt.Errorf("Argument 3 (build_args): %v", err)
		}
	}

	bs, err := s.readFile(dockerfilePath)
	if err != nil {
		return nil, err
	}

	if s.imagesByName[ref.Name()] != nil {
		return nil, fmt.Errorf("Image for ref %q has already been defined", ref.Name())
	}

	cachePaths, err := s.cachePathsFromSkylarkValue(cacheVal)
	if err != nil {
		return nil, err
	}

	// Now we have everything we need, and construct it into the final form

	dockerInfo := model.DockerInfo{
		DockerRef: ref,
		Details: model.StaticBuild{
			Dockerfile: string(bs),
			BuildPath:  string(context.path),
			BuildArgs:  sba,
		},
	}.WithCachePaths(cachePaths)

	ignores := newIgnores()
	ignores.add(context.path)

	repos := newRepos()
	repos.add(context.repo)

	r := &staticBuild{
		dockerInfo: dockerInfo,
		ignores:    ignores.lst,
		repos:      repos.lst,
	}
	s.imagesByName[ref.Name()] = r
	s.images = append(s.images, r)

	return skylark.None, nil
}

func (s *tiltfileState) fastBuild(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {

	var dockerRef, entrypoint string
	var baseDockerfile skylark.Value
	var cacheVal skylark.Value
	err := skylark.UnpackArgs(fn.Name(), args, kwargs,
		"ref", &dockerRef,
		"base_dockerfile", &baseDockerfile,
		"entrypoint?", &entrypoint,
		"cache?", &cacheVal,
	)
	if err != nil {
		return nil, err
	}

	baseDockerfilePath, err := s.localPathFromSkylarkValue(baseDockerfile)
	if err != nil {
		return nil, fmt.Errorf("Argument 2 (base_dockerfile): %v", err)
	}

	ref, err := reference.ParseNormalizedNamed(dockerRef)
	if err != nil {
		return nil, fmt.Errorf("Parsing %q: %v", dockerRef, err)
	}

	if s.imagesByName[ref.Name()] != nil {
		return nil, fmt.Errorf("Image for ref %q has already been defined", ref.Name())
	}

	bs, err := s.readFile(baseDockerfilePath)
	if err != nil {
		return nil, err
	}

	df := dockerfile.Dockerfile(bs)
	if err = df.ValidateBaseDockerfile(); err != nil {
		return nil, err
	}

	cachePaths, err := s.cachePathsFromSkylarkValue(cacheVal)
	if err != nil {
		return nil, err
	}

	r := &fastBuild{
		s:              s,
		ref:            ref,
		baseDockerfile: string(df),
		entrypoint:     entrypoint,
		cachePaths:     cachePaths,
	}
	s.imagesByName[ref.Name()] = r
	s.images = append(s.images, r)

	return r, nil
}

func (s *tiltfileState) cachePathsFromSkylarkValue(val skylark.Value) ([]string, error) {
	if val == nil {
		return nil, nil
	}
	if val, ok := val.(skylark.Sequence); ok {
		var result []string
		it := val.Iterate()
		defer it.Done()
		var i skylark.Value
		for it.Next(&i) {
			str, ok := i.(skylark.String)
			if !ok {
				return nil, fmt.Errorf("cache param %v is a %T; must be a string", i, i)
			}
			result = append(result, string(str))
		}
		return result, nil
	}
	str, ok := val.(skylark.String)
	if !ok {
		return nil, fmt.Errorf("cache param %v is a %T; must be a string or a sequence of strings", val, val)
	}
	return []string{string(str)}, nil
}

type fastBuild struct {
	s              *tiltfileState
	ref            reference.Named
	baseDockerfile string
	entrypoint     string
	cachePaths     []string
	mounts         []mount
	steps          []model.Step
}

var _ skylark.Value = &fastBuild{}
var _ dockerImage = &fastBuild{}

func (b *fastBuild) String() string {
	return fmt.Sprintf("fast_build(%q)", b.ref.Name())
}

func (b *fastBuild) Type() string {
	return "fast_build"
}

func (b *fastBuild) Freeze() {}

func (b *fastBuild) Truth() skylark.Bool {
	return true
}

func (b *fastBuild) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: fast_build")
}

func (b *fastBuild) toDomain() (model.BuildInfo, []model.Dockerignore, []model.LocalGithubRepo) {
	ignores := newIgnores()
	repos := newRepos()
	var mounts []model.Mount
	for _, m := range b.mounts {
		ignores.add(m.src.path)
		repos.add(m.src.repo)
		if m.src.repo != nil {
			ignores.add(m.src.repo.basePath)
		}
		mounts = append(mounts, model.Mount{LocalPath: m.src.path, ContainerPath: m.mountPoint})
	}

	// FIXME(dbentley): do we need others? e.g. the base dockerfile path?

	info := model.DockerInfo{
		DockerRef: b.ref,
		Details: model.FastBuild{
			Entrypoint:     model.ToShellCmd(b.entrypoint),
			BaseDockerfile: b.baseDockerfile,
			Mounts:         mounts,
			Steps:          b.steps,
		},
	}.WithCachePaths(b.cachePaths)

	return info, ignores.lst, repos.lst
}

const (
	addN = "add"
	runN = "run"
)

func (b *fastBuild) Attr(name string) (skylark.Value, error) {
	switch name {
	case addN:
		return skylark.NewBuiltin(name, b.add), nil
	case runN:
		return skylark.NewBuiltin(name, b.run), nil
	default:
		return skylark.None, nil
	}
}

func (b *fastBuild) AttrNames() []string {
	return []string{addN, runN}
}

func (b *fastBuild) add(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	if len(b.steps) > 0 {
		return nil, fmt.Errorf("fast_build(%q).add() called after .run(); must add all code before runs", b.ref.Name())
	}

	var src skylark.Value
	var mountPoint string

	if err := skylark.UnpackArgs(fn.Name(), args, kwargs, "src", &src, "dest", &mountPoint); err != nil {
		return nil, err
	}

	m := mount{}
	switch p := src.(type) {
	case localPath:
		m.src = p
	case *gitRepo:
		m.src = p.makeLocalPath("")
	default:
		return nil, fmt.Errorf("fast_build(%q).add(): invalid type for src. Got %s want gitRepo OR localPath", fn.Name(), src.Type())
	}

	m.mountPoint = mountPoint
	b.mounts = append(b.mounts, m)

	return b, nil
}

func (b *fastBuild) run(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var cmd string
	var trigger skylark.Value
	if err := skylark.UnpackArgs(fn.Name(), args, kwargs, "cmd", &cmd, "trigger?", &trigger); err != nil {
		return nil, err
	}

	var triggers []string
	switch trigger := trigger.(type) {
	case *skylark.List:
		l := trigger.Len()
		triggers = make([]string, l)
		for i := 0; i < l; i++ {
			t := trigger.Index(i)
			tStr, isStr := t.(skylark.String)
			if !isStr {
				return nil, badTypeErr(fn, skylark.String(""), t)
			}
			triggers[i] = string(tStr)
		}
	case skylark.String:
		triggers = []string{string(trigger)}
	}

	step := model.ToStep(b.s.absWorkingDir(), model.ToShellCmd(cmd))
	step.Triggers = triggers

	b.steps = append(b.steps, step)
	return b, nil
}

type mount struct {
	src        localPath
	mountPoint string
}

func mountsToDomain(mounts []mount) []model.Mount {
	var result []model.Mount

	for _, m := range mounts {
		result = append(result, model.Mount{LocalPath: m.src.path, ContainerPath: m.mountPoint})
	}

	return result
}

type repos struct {
	lst []model.LocalGithubRepo
	set map[string]bool
}

func newRepos() *repos {
	return &repos{set: map[string]bool{}}
}

func (r *repos) add(repo *gitRepo) {
	if repo == nil || r.set[repo.basePath] {
		return
	}

	r.set[repo.basePath] = true
	r.lst = append(r.lst, model.LocalGithubRepo{
		LocalPath:         repo.basePath,
		GitignoreContents: repo.gitignoreContents,
	})
}

type ignores struct {
	lst []model.Dockerignore
	set map[string]bool
}

func newIgnores() *ignores {
	return &ignores{set: map[string]bool{}}
}

func (i *ignores) add(p string) {
	if p == "" || i.set[p] || !ospath.IsDir(p) {
		return
	}
	i.set[p] = true

	contents, err := ioutil.ReadFile(filepath.Join(p, ".dockerignore"))
	if err != nil {
		return
	}

	i.lst = append(i.lst, model.Dockerignore{
		LocalPath: p,
		Contents:  string(contents),
	})

}
