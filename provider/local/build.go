package local

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

const (
	BuildCacheDuration = 5 * time.Minute
)

var buildUpdateLock sync.Mutex

func (p *Provider) BuildCreate(app, url string, opts types.BuildCreateOptions) (*types.Build, error) {
	log := p.logger("BuildCreate").Append("app=%q url=%q", app, url)

	a, err := p.AppGet(app)
	if err != nil {
		return nil, log.Error(err)
	}

	id := types.Id("B", 10)

	b := &types.Build{
		Id:      id,
		App:     app,
		Status:  "created",
		Created: time.Now().UTC(),
	}

	if err := p.storageStore(fmt.Sprintf("apps/%s/builds/%s", app, id), b); err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	registries, err := p.RegistryList()
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	auth, err := json.Marshal(registries)
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	sys, err := p.SystemGet()
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	buildUpdateLock.Lock()
	defer buildUpdateLock.Unlock()

	pid, err := p.ProcessStart(app, types.ProcessRunOptions{
		Command: fmt.Sprintf("build -id %s -url %s", id, url),
		Environment: map[string]string{
			"BUILD_APP":         app,
			"BUILD_AUTH":        base64.StdEncoding.EncodeToString(auth),
			"BUILD_DEVELOPMENT": fmt.Sprintf("%t", opts.Development),
			"BUILD_PREFIX":      fmt.Sprintf("%s/%s", p.Name, app),
		},
		Name:    fmt.Sprintf("%s-build-%s", app, id),
		Image:   sys.Image,
		Release: a.Release,
		Service: "build",
		Volumes: map[string]string{
			"/var/run/docker.sock": "/var/run/docker.sock",
		},
	})
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	b, err = p.BuildGet(app, id)
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	b.Process = pid

	if err := p.storageStore(fmt.Sprintf("apps/%s/builds/%s", app, id), b); err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	return b, log.Successf("id=%s", b.Id)
}

func (p *Provider) BuildGet(app, id string) (*types.Build, error) {
	log := p.logger("BuildGet").Append("app=%q id=%q", app, id)

	var b *types.Build

	if err := p.storageLoad(fmt.Sprintf("apps/%s/builds/%s", app, id), &b, BuildCacheDuration); err != nil {
		if strings.HasPrefix(err.Error(), "no such key:") {
			return nil, log.Error(fmt.Errorf("no such build: %s", id))
		} else {
			return nil, errors.WithStack(log.Error(err))
		}
	}

	return b, log.Success()
}

func (p *Provider) BuildList(app string) (types.Builds, error) {
	log := p.logger("BuildList").Append("app=%q", app)

	ids, err := p.storageList(fmt.Sprintf("apps/%s/builds", app))
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	builds := make(types.Builds, len(ids))

	for i, id := range ids {
		build, err := p.BuildGet(app, id)
		if err != nil {
			return nil, errors.WithStack(log.Error(err))
		}

		builds[i] = *build
	}

	sort.Slice(builds, func(i, j int) bool { return builds[i].Created.Before(builds[j].Created) })

	return builds, log.Success()
}

func (p *Provider) BuildLogs(app, id string) (io.ReadCloser, error) {
	log := p.logger("BuildLogs").Append("app=%q id=%q", app, id)

	build, err := p.BuildGet(app, id)
	if err != nil {
		return nil, log.Error(err)
	}

	switch build.Status {
	case "running":
		log.Success()
		return p.ProcessLogs(app, build.Process, types.LogsOptions{Follow: true, Prefix: false})
	default:
		log.Success()
		return p.ObjectFetch(app, fmt.Sprintf("convox/builds/%s/log", id))
	}
}

func (p *Provider) BuildUpdate(app, id string, opts types.BuildUpdateOptions) (*types.Build, error) {
	buildUpdateLock.Lock()
	defer buildUpdateLock.Unlock()

	log := p.logger("BuildUpdate").Append("app=%q id=%q", app, id)

	build, err := p.BuildGet(app, id)
	if err != nil {
		return nil, log.Error(err)
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

	if err := p.storageStore(fmt.Sprintf("apps/%s/builds/%s", app, id), build); err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	return build, log.Success()
}
