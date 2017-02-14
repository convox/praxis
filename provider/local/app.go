package local

import (
	"fmt"

	"github.com/convox/praxis/provider/types"
)

func (p *Provider) AppCreate(name string) (*types.App, error) {
	app := &types.App{
		Name: name,
	}

	if err := p.Store(fmt.Sprintf("apps/%s/app", app.Name), app); err != nil {
		return nil, err
	}

	return app, nil
}

func (p *Provider) AppDelete(app string) error {
	return p.DeleteAll(fmt.Sprintf("apps/%s", app))
}
