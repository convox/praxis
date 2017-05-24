package local

import (
	"fmt"
	"sort"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
)

func (p *Provider) ResourceGet(app, name string) (*types.Resource, error) {
	rs, err := p.ResourceList(app)
	if err != nil {
		return nil, err
	}

	for _, r := range rs {
		if r.Name == name {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("resource not found: %s", name)
}

func (p *Provider) ResourceList(app string) (types.Resources, error) {
	m, _, err := helpers.AppManifest(p, app)
	if err != nil {
		return nil, err
	}

	rs := make(types.Resources, len(m.Resources))

	for i, r := range m.Resources {
		e, err := resourceURL(app, r.Type, r.Name)
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
