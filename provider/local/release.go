package local

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) ReleaseCreate(app string, opts types.ReleaseCreateOptions) (*types.Release, error) {
	id := types.Id("R", 10)

	release := &types.Release{
		Id:    id,
		App:   app,
		Build: opts.Build,
		Env:   opts.Env,
	}

	if err := p.Store(fmt.Sprintf("apps/%s/releases/%s", app, id), release); err != nil {
		return nil, err
	}

	return release, nil
}

func (p *Provider) ReleaseGet(app, id string) (release *types.Release, err error) {
	err = p.Load(fmt.Sprintf("/apps/%s/releases/%s", app, id), &release)
	return
}
