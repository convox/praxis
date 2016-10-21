package local

import "github.com/convox/praxis/provider"

func (p *Provider) AppCreate(name string, opts provider.AppCreateOptions) (*provider.App, error) {
	err := p.TableSave("system", "apps", name, nil)
	if err != nil {
		return nil, err
	}

	app := &provider.App{
		Name: name,
	}

	return app, nil
}

func (p *Provider) AppDelete(name string) error {
	return p.TableRemove("system", "apps", name)
}
