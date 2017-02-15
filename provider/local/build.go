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

	pid, err := p.Run(app, "build", "convox/build", "build", "-id", id, "-url", url)
	if err != nil {
		return nil, err
	}

	fmt.Printf("pid = %+v\n", pid)

	if err := p.Store(fmt.Sprintf("apps/%s/builds/%s", app, id), build); err != nil {
		return nil, err
	}

	build.Release = types.Id("R", 10)

	return build, nil
}
