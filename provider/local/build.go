package local

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) BuildCreate(app, url string, opts types.BuildCreateOptions) (*types.Build, error) {
	id := types.Id("B", 10)

	build := &types.Build{
		Id: id,
	}

	// pid, err := p.Run(app, "build", "convox/build", "build", "-id", id, "-url", url)

	if err := p.Store(fmt.Sprintf("apps/%s/builds/%s", app, id), build); err != nil {
		return nil, err
	}

	return build, nil
}
