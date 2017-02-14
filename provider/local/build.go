package local

import "github.com/convox/praxis/provider/types"

func (p *Provider) BuildCreate(app, url string, opts types.BuildCreateOptions) (*types.Build, error) {
	return &types.Build{Id: "B1234", Release: "R1234"}, nil
}
