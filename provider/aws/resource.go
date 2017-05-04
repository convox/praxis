package aws

import (
	"fmt"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func (p *Provider) ResourceList(app string) (types.Resources, error) {
	a, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}
	if a.Release == "" {
		return nil, fmt.Errorf("no release for app: %s\n", app)
	}

	r, err := p.ReleaseGet(app, a.Release)
	if err != nil {
		return nil, err
	}
	if r.Build == "" {
		return nil, fmt.Errorf("no build for app: %s\n", app)
	}

	b, err := p.BuildGet(app, r.Build)
	if err != nil {
		return nil, err
	}

	m, err := manifest.Load([]byte(b.Manifest))
	if err != nil {
		return nil, err
	}

	rs := make(types.Resources, len(m.Resources))

	for i, r := range m.Resources {
		stack, err := p.appResource(app, fmt.Sprintf("Resource%s", upperName(r.Name)))
		if err != nil {
			return nil, err
		}

		url, err := p.stackOutput(stack, "Url")
		if err != nil {
			return nil, err
		}

		fmt.Printf("i = %+v\n", i)
		fmt.Printf("stack = %+v\n", stack)
		fmt.Printf("url = %+v\n", url)

		// rs[i] = types.Resource{
		//   Name:     r.Name,
		//   Endpoint: url,
		//   Type:     r.Type,
		// }
	}

	fmt.Printf("rs = %+v\n", rs)

	return nil, fmt.Errorf("unimplemented")
}
