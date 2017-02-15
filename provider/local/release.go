package local

import "github.com/convox/praxis/types"

func (p *Provider) ReleaseGet(app, id string) (*types.Release, error) {
	release := &types.Release{
		Id:  id,
		App: app,
	}

	return release, nil
}
