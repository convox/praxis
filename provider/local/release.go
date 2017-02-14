package local

import "github.com/convox/praxis/provider/types"

func (p *Provider) ReleaseGet(app, id string) (*types.Release, error) {
	return &types.Release{Id: "R1234", Build: "B1234"}, nil
}
