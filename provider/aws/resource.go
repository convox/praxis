package aws

import (
	"fmt"
	"sort"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
)

func (p *Provider) ResourceList(app string) (types.Resources, error) {
	m, _, err := helpers.AppManifest(p, app)
	if err != nil {
		return nil, err
	}

	rs := make(types.Resources, len(m.Resources))

	for i, r := range m.Resources {
		stack, err := p.appResource(app, fmt.Sprintf("Resource%s", upperName(r.Name)))
		if err != nil {
			return nil, err
		}

		e, err := p.stackOutput(stack, "Url")
		if err != nil {
			return nil, err
		}

		rs[i] = types.Resource{
			Name:     r.Name,
			Endpoint: e,
			Type:     r.Type,
		}
	}

	sort.Slice(rs, func(i, j int) bool { return rs[i].Name < rs[j].Name })

	return rs, nil
}
