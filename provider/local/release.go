package local

import (
	"fmt"
	"sort"
	"time"

	"github.com/convox/praxis/types"
)

func (p *Provider) ReleaseCreate(app string, opts types.ReleaseCreateOptions) (*types.Release, error) {
	id := types.Id("R", 10)

	release := &types.Release{
		Id:      id,
		App:     app,
		Build:   opts.Build,
		Env:     opts.Env,
		Created: time.Now(),
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

func (p *Provider) ReleaseList(app string) (types.Releases, error) {
	ids, err := p.List(fmt.Sprintf("apps/%s/releases", app))
	if err != nil {
		return nil, err
	}

	releases := make(types.Releases, len(ids))

	for i, id := range ids {
		release, err := p.ReleaseGet(app, id)
		if err != nil {
			return nil, err
		}

		releases[i] = *release
	}

	sort.Sort(releases)

	return releases, nil
}
