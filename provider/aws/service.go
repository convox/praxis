package aws

import (
	"fmt"
	"strings"

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
		endpoint := ""

		o, err := p.appOutput(app, fmt.Sprintf("Endpoint%s", upperName(s.Name)))
		if err != nil && !strings.Contains(err.Error(), "no such output for stack") {
			return nil, err
		}

		if o != "" {
			endpoint = fmt.Sprintf("https://%s", o)
		}

		ss = append(ss, types.Service{
			Name:     s.Name,
			Endpoint: endpoint,
		})
	}

	return ss, nil
}
