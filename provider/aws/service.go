package aws

import (
	"fmt"

	"github.com/convox/praxis/cache"
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
	// clear the AppGet cache to avoid "no releases for app"
	err := cache.Clear("describeStack", fmt.Sprintf("%s-%s", p.Name, app))
	if err != nil {
		return nil, err
	}

	m, _, err := helpers.AppManifest(p, app)
	if err != nil {
		return nil, err
	}

	ss := types.Services{}

	for _, s := range m.Services {
		endpoint, err := p.appOutput(app, fmt.Sprintf("Endpoint%s", upperName(s.Name)))
		if err != nil {
			return nil, err
		}

		ss = append(ss, types.Service{
			Name:     s.Name,
			Endpoint: fmt.Sprintf("https://%s", endpoint),
		})
	}

	return ss, nil
}
