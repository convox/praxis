package local

import (
	"fmt"
	"sort"
	"time"

	"github.com/convox/praxis/manifest"
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

func (p *Provider) ReleasePromote(app, id string) error {
	a, err := p.AppGet(app)
	if err != nil {
		return err
	}

	r, err := p.ReleaseGet(app, id)
	if err != nil {
		return err
	}

	b, err := p.BuildGet(app, r.Build)
	if err != nil {
		return err
	}

	a.Release = r.Id

	if err := p.Store(fmt.Sprintf("apps/%s/app.json", a.Name), a); err != nil {
		return err
	}

	m, err := manifest.Load([]byte(b.Manifest))
	if err != nil {
		return err
	}

	for _, t := range m.Tables {
		if err := p.TableCreate(app, t.Name, types.TableCreateOptions{Indexes: t.Indexes}); err != nil {
			return err
		}
	}

	for _, t := range m.Timers {
		opts := types.TimerCreateOptions{
			Command:  t.Command,
			Schedule: t.Schedule,
			Service:  t.Service,
		}

		if err := p.TimerCreate(app, t.Name, opts); err != nil {
			return err
		}
	}

	return nil
}
