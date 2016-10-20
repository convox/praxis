package local

import "github.com/convox/praxis/provider/models"

func (p *Provider) AppCreate(name string, opts models.AppCreateOptions) (*models.App, error) {
	err := p.TableSave("system", "apps", name, nil)
	if err != nil {
		return nil, err
	}

	app := &models.App{
		Name: name,
	}

	return app, nil
}
