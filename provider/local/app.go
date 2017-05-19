package local

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

func (p *Provider) AppCreate(name string) (*types.App, error) {
	log := p.logger("AppCreate").Append("name=%q", name)

	if p.storageExists(fmt.Sprintf("apps/%s/app.json", name)) {
		return nil, log.Error(fmt.Errorf("app already exists: %s", name))
	}

	app := &types.App{
		Name:   name,
		Status: "running",
	}

	if err := p.storageStore(fmt.Sprintf("apps/%s/app.json", app.Name), app); err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	return app, log.Success()
}

func (p *Provider) AppDelete(app string) error {
	log := p.logger("AppDelete").Append("app=%q", app)

	if _, err := p.AppGet(app); err != nil {
		return log.Error(err)
	}

	pss, err := p.ProcessList(app, types.ProcessListOptions{})
	if err != nil {
		return errors.WithStack(log.Error(err))
	}

	for _, ps := range pss {
		if err := p.ProcessStop(app, ps.Id); err != nil {
			return errors.WithStack(log.Error(err))
		}
	}

	if err := p.storageDeleteAll(fmt.Sprintf("apps/%s", app)); err != nil {
		return errors.WithStack(log.Error(err))
	}

	return log.Success()
}

func (p *Provider) AppGet(name string) (*types.App, error) {
	log := p.logger("AppGet").Append("name=%q", name)

	var app types.App

	if err := p.storageLoad(fmt.Sprintf("apps/%s/app.json", name), &app); err != nil {
		if strings.HasPrefix(err.Error(), "no such key:") {
			return nil, log.Error(api.Errorf(404, "no such app: %s", name))
		} else {
			return nil, errors.WithStack(log.Error(err))
		}
	}

	return &app, log.Success()
}

func (p *Provider) AppList() (types.Apps, error) {
	log := p.logger("AppList")

	names, err := p.storageList("apps")
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	apps := make(types.Apps, len(names))

	for i, name := range names {
		app, err := p.AppGet(name)
		if err != nil {
			return nil, errors.WithStack(log.Error(err))
		}

		apps[i] = *app
	}

	sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })

	return apps, log.Successf("count=%d", len(apps))
}

func (p *Provider) AppLogs(app string, opts types.LogsOptions) (io.ReadCloser, error) {
	log := p.logger("AppLogs").Append("app=%q", app)

	if _, err := p.AppGet(app); err != nil {
		return nil, log.Error(err)
	}

	pss, err := p.ProcessList(app, types.ProcessListOptions{})
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	r, w := io.Pipe()

	for _, ps := range pss {
		go p.processLogs(app, ps, opts, w)
	}

	return r, log.Success()
}

func (p *Provider) AppRegistry(app string) (*types.Registry, error) {
	log := p.logger("AppRegistry").Append("app=%q", app)

	if _, err := p.AppGet(app); err != nil {
		return nil, log.Error(err)
	}

	registry := &types.Registry{
		Hostname: p.Name,
		Username: "",
		Password: "",
	}

	return registry, log.Success()
}
