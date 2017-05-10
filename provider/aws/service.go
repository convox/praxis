package aws

import (
	"fmt"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
)

func (p *Provider) ServiceList(app string) (types.Services, error) {
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
			Endpoint: endpoint,
		})
	}

	return ss, nil
}
