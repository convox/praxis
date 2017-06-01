package local

import (
	"fmt"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

func (p *Provider) ServiceList(app string) (types.Services, error) {
	log := p.logger("ServiceList").Append("app=%q", app)

	m, _, err := helpers.AppManifest(p, app)
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
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

	return ss, log.Success()
}
