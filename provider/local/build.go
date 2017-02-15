package local

import (
	"fmt"

	"github.com/convox/praxis/provider/types"
)

func (p *Provider) BuildCreate(app, url string, opts types.BuildCreateOptions) (*types.Build, error) {
	id := types.Id("B", 10)

	build := &types.Build{
		Id: id,
	}

	if err := p.Store(fmt.Sprintf("apps/%s/builds/%s", app, id), build); err != nil {
		return nil, err
	}

	return build, nil
}
