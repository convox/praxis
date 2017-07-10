package aws

import (
	"fmt"
	"io"
	"sort"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

func (p *Provider) ResourceGet(app, name string) (*types.Resource, error) {
	m, _, err := helpers.AppManifest(p, app)
	if err != nil {
		return nil, err
	}

	for _, r := range m.Resources {
		if r.Name == name {
			return p.resourceFromManifest(app, r)
		}
	}

	return nil, fmt.Errorf("resource not found: %s", name)
}

func (p *Provider) ResourceList(app string) (types.Resources, error) {
	m, _, err := helpers.AppManifest(p, app)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	rs := make(types.Resources, len(m.Resources))

	for i, r := range m.Resources {
		rr, err := p.resourceFromManifest(app, r)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		rs[i] = *rr
	}

	sort.Slice(rs, func(i, j int) bool { return rs[i].Name < rs[j].Name })

	return rs, nil
}

func (p *Provider) ResourceProxy(app, resource string, in io.Reader) (io.ReadCloser, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) resourceFromManifest(app string, r manifest.Resource) (*types.Resource, error) {
	stack, err := p.appResource(app, fmt.Sprintf("Resource%s", upperName(r.Name)))
	if err != nil {
		return nil, err
	}

	e, err := p.stackOutput(stack, "Url")
	if err != nil {
		return nil, err
	}

	rr := &types.Resource{
		Name:     r.Name,
		Endpoint: e,
		Type:     r.Type,
	}

	return rr, nil
}
