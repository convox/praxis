package aws

import (
	"fmt"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func (p *Provider) ServiceList(app string) (types.Services, error) {
	domain, err := p.rackOutput("Domain")
	if err != nil {
		return nil, err
	}

	a, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	if a.Release == "" {
		return types.Services{}, nil
	}

	r, err := p.ReleaseGet(app, a.Release)
	if err != nil {
		return nil, err
	}

	b, err := p.BuildGet(app, r.Build)
	if err != nil {
		return nil, err
	}

	m, err := manifest.Load([]byte(b.Manifest))
	if err != nil {
		return nil, err
	}

	ss := types.Services{}

	for _, s := range m.Services {
		endpoint := ""

		if s.Port.Port > 0 {
			endpoint = fmt.Sprintf("https://%s-%s.%s", app, s.Name, domain)
		}

		ss = append(ss, types.Service{
			Name:     s.Name,
			Endpoint: endpoint,
		})
	}

	return ss, nil
}
