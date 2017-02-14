package local

import "github.com/convox/praxis/provider/types"

func (p *Provider) AppCreate(name string) (*types.App, error) {
	return &types.App{Name: name}, nil
}
