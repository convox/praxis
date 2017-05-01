package local

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/convox/praxis/types"
)

func (p *Provider) AppCreate(name string) (*types.App, error) {
	if _, err := p.AppGet(name); err == nil {
		return nil, fmt.Errorf("app already exists: %s", name)
	}

	app := &types.App{
		Name:   name,
		Status: "running",
	}

	if err := p.storageStore(fmt.Sprintf("apps/%s/app.json", app.Name), app); err != nil {
		return nil, err
	}

	return app, nil
}

func (p *Provider) AppDelete(app string) error {
	pss, err := p.ProcessList(app, types.ProcessListOptions{})
	if err != nil {
		return err
	}

	for _, ps := range pss {
		if err := p.ProcessStop(app, ps.Id); err != nil {
			return err
		}
	}

	if _, err := p.AppGet(app); err != nil {
		return err
	}

	return p.storageDeleteAll(fmt.Sprintf("apps/%s", app))
}

func (p *Provider) AppGet(name string) (*types.App, error) {
	var app types.App

	if err := p.storageLoad(fmt.Sprintf("apps/%s/app.json", name), &app); err != nil {
		if strings.HasPrefix(err.Error(), "no such key:") {
			return nil, fmt.Errorf("no such app: %s", name)
		}
		return nil, err
	}

	return &app, nil
}

func (p *Provider) AppList() (types.Apps, error) {
	names, err := p.storageList("apps")
	if err != nil {
		return nil, err
	}

	apps := make(types.Apps, len(names))

	for i, name := range names {
		app, err := p.AppGet(name)
		if err != nil {
			return nil, err
		}

		apps[i] = *app
	}

	sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })

	return apps, nil
}

func (p *Provider) AppLogs(app string, opts types.AppLogsOptions) (io.ReadCloser, error) {
	pss, err := p.ProcessList(app, types.ProcessListOptions{})
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	for _, ps := range pss {
		go p.processLogs(app, ps, opts, w)
	}

	return r, nil
}

func (p *Provider) AppRegistry(app string) (*types.Registry, error) {
	registry := &types.Registry{
		Hostname: p.Name,
		Username: "",
		Password: "",
	}

	return registry, nil
}

func (p *Provider) processLogs(app string, ps types.Process, opts types.AppLogsOptions, w io.Writer) {
	// TODO: use opts

	r, err := p.ProcessLogs(app, ps.Id)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		return
	}

	s := bufio.NewScanner(r)

	for s.Scan() {
		w.Write([]byte(fmt.Sprintf("[%s.%s] %s\n", ps.Service, ps.Id, s.Text())))
	}
}
