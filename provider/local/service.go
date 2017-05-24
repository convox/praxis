package local

import (
	"fmt"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
)

func (p *Provider) ServiceGet(app, name string) (*types.Service, error) {
	ss, err := p.ServiceList(app)
	if err != nil {
		return nil, err
	}

	for _, s := range ss {
		if s.Name == name {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("service not found: %s", name)
}

func (p *Provider) ServiceList(app string) (types.Services, error) {
	m, _, err := helpers.AppManifest(p, app)
	if err != nil {
		return nil, err
	}

	ss := types.Services{}

	for _, s := range m.Services {
		endpoint := ""

		if s.Port.Port > 0 {
			endpoint = fmt.Sprintf("https://%s.%s.%s", s.Name, app, p.Name)
		}

		ss = append(ss, types.Service{
			Name:     s.Name,
			Endpoint: endpoint,
		})
	}

	return ss, nil
}
