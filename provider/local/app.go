package local

import (
	"fmt"
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

	if err := p.Store(fmt.Sprintf("apps/%s/app.json", app.Name), app); err != nil {
		return nil, err
	}

	return app, nil
}

func (p *Provider) AppDelete(name string) error {
	if _, err := p.AppGet(name); err != nil {
		return err
	}

	return p.DeleteAll(fmt.Sprintf("apps/%s", name))
}

func (p *Provider) AppGet(name string) (*types.App, error) {
	var app types.App

	if err := p.Load(fmt.Sprintf("apps/%s/app.json", name), &app); err != nil {
		if strings.HasPrefix(err.Error(), "no such key:") {
			return nil, fmt.Errorf("no such app: %s", name)
		}
		return nil, err
	}

	return &app, nil
}

func (p *Provider) AppList() (types.Apps, error) {
	names, err := p.List("apps")
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
