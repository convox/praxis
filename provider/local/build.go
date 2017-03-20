package local

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/convox/praxis/types"
)

func (p *Provider) BuildCreate(app, url string, opts types.BuildCreateOptions) (*types.Build, error) {
	_, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	id := types.Id("B", 10)

	build := &types.Build{
		Id:      id,
		App:     app,
		Status:  "created",
		Created: time.Now(),
	}

	if err := p.Store(fmt.Sprintf("apps/%s/builds/%s", app, id), build); err != nil {
		return nil, err
	}

	registries, err := p.RegistryList()
	if err != nil {
		return nil, err
	}

	auth, err := json.Marshal(registries)
	if err != nil {
		return nil, err
	}

	pid, err := p.ProcessStart(app, types.ProcessRunOptions{
		Command: fmt.Sprintf("build -id %s -url %s", id, url),
		Environment: map[string]string{
			"BUILD_APP":  app,
			"BUILD_AUTH": base64.StdEncoding.EncodeToString(auth),
		},
		Name:    fmt.Sprintf("%s-build-%s", app, id),
		Image:   "convox/praxis:test8",
		Service: "build",
		Volumes: map[string]string{
			"/var/run/docker.sock": "/var/run/docker.sock",
		},
	})
	if err != nil {
		return nil, err
	}

	build, err = p.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	build.Process = pid

	if err := p.Store(fmt.Sprintf("apps/%s/builds/%s", app, id), build); err != nil {
		return nil, err
	}

	return build, nil
}

func (p *Provider) BuildGet(app, id string) (build *types.Build, err error) {
	err = p.Load(fmt.Sprintf("apps/%s/builds/%s", app, id), &build)
	return
}

func (p *Provider) BuildList(app string) (types.Builds, error) {
	ids, err := p.List(fmt.Sprintf("apps/%s/builds", app))
	if err != nil {
		return nil, err
	}

	builds := make(types.Builds, len(ids))

	for i, id := range ids {
		build, err := p.BuildGet(app, id)
		if err != nil {
			return nil, err
		}

		builds[i] = *build
	}

	sort.Slice(builds, func(i, j int) bool { return builds[i].Created.Before(builds[j].Created) })

	return builds, nil
}

func (p *Provider) BuildLogs(app, id string) (io.ReadCloser, error) {
	build, err := p.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	switch build.Status {
	case "running":
		return p.ProcessLogs(app, build.Process)
	default:
		return p.ObjectFetch(app, fmt.Sprintf("convox/builds/%s/log", id))
	}
}

func (p *Provider) BuildUpdate(app, id string, opts types.BuildUpdateOptions) (*types.Build, error) {
	build, err := p.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	if !opts.Ended.IsZero() {
		build.Ended = opts.Ended
	}

	if opts.Manifest != "" {
		build.Manifest = opts.Manifest
	}

	if opts.Release != "" {
		build.Release = opts.Release
	}

	if !opts.Started.IsZero() {
		build.Started = opts.Started
	}

	if opts.Status != "" {
		build.Status = opts.Status
	}

	if err := p.Store(fmt.Sprintf("apps/%s/builds/%s", app, id), build); err != nil {
		return nil, err
	}

	return build, nil
}
