package local

import "github.com/convox/praxis/types"

func (p *Provider) SystemGet() (*types.System, error) {
	system := &types.System{
		Name:    "convox",
		Version: "dev",
	}

	return system, nil
}
