package local

import (
	"fmt"
	"io"
	"math/rand"
	"sort"
	"time"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (p *Provider) ReleaseCreate(app string, opts types.ReleaseCreateOptions) (*types.Release, error) {
	r, err := p.releaseFork(app)
	if err != nil {
		return nil, err
	}

	if opts.Build != "" {
		r.Build = opts.Build
	}

	if opts.Env != nil {
		r.Env = opts.Env
	}

	r.Stage = opts.Stage

	if err := p.storageStore(fmt.Sprintf("apps/%s/releases/%s/release.json", app, r.Id), r); err != nil {
		return nil, err
	}

	return r, nil
}

func (p *Provider) ReleaseGet(app, id string) (*types.Release, error) {
	a, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	var r *types.Release

	if err := p.storageLoad(fmt.Sprintf("apps/%s/releases/%s/release.json", app, id), r); err != nil {
		return nil, err
	}

	if r.Env == nil {
		r.Env = types.Environment{}
	}

	if r.Id == a.Release {
		r.Status = "current"
	}

	return r, nil
}

func (p *Provider) ReleaseList(app string, opts types.ReleaseListOptions) (types.Releases, error) {
	a, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	ids, err := p.storageList(fmt.Sprintf("apps/%s/releases", app))
	if err != nil {
		return nil, err
	}

	releases := make(types.Releases, len(ids))

	for i, id := range ids {
		release, err := p.ReleaseGet(app, id)
		if err != nil {
			return nil, err
		}

		if release.Id == a.Release {
			release.Status = "current"
		}

		releases[i] = *release
	}

	sort.Slice(releases, func(i, j int) bool { return releases[j].Created.Before(releases[i].Created) })

	limit := coalescei(opts.Count, 10)

	if len(releases) > limit {
		releases = releases[0:limit]
	}

	return releases, nil
}

func (p *Provider) ReleaseLogs(app, id string, opts types.LogsOptions) (io.ReadCloser, error) {
	key := fmt.Sprintf("apps/%s/releases/%s/log", app, id)

	r, err := p.ReleaseGet(app, id)
	if err != nil {
		return nil, err
	}

	for {
		if r.Status != "created" {
			break
		}

		r, err = p.ReleaseGet(app, id)
		if err != nil {
			return nil, err
		}

		time.Sleep(1 * time.Second)
	}

	lr, lw := io.Pipe()

	go func() {
		defer lw.Close()

		var since time.Time

		for {
			time.Sleep(200 * time.Millisecond)

			p.storageLogRead(key, since, func(at time.Time, entry []byte) {
				since = at
				lw.Write(entry)
			})

			if !opts.Follow {
				break
			}

			r, err := p.ReleaseGet(app, id)
			if err != nil {
				continue
			}

			if r.Status == "promoted" || r.Status == "failed" {
				break
			}
		}
	}()

	return lr, nil
}

func (p *Provider) ReleasePromote(app, release string) error {
	a, err := p.AppGet(app)
	if err != nil {
		return err
	}

	r, err := p.ReleaseGet(app, release)
	if err != nil {
		return err
	}

	if r.Build == "" {
		return nil
	}

	r.Status = "running"

	if err := p.storageStore(fmt.Sprintf("apps/%s/releases/%s/release.json", app, release), r); err != nil {
		return err
	}

	b, err := p.BuildGet(app, r.Build)
	if err != nil {
		return err
	}

	a.Release = r.Id

	if err := p.storageStore(fmt.Sprintf("apps/%s/app.json", a.Name), a); err != nil {
		return err
	}

	m, err := manifest.Load([]byte(b.Manifest), r.Env)
	if err != nil {
		return err
	}

	if err := p.converge(app); err != nil {
		return err
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

	r.Status = "promoted"

	if err := p.storageStore(fmt.Sprintf("apps/%s/releases/%s/release.json", app, release), r); err != nil {
		return err
	}

	return nil
}

func (p *Provider) releaseFork(app string) (*types.Release, error) {
	r := &types.Release{
		Id:      types.Id("R", 10),
		App:     app,
		Status:  "created",
		Created: time.Now().UTC(),
	}

	rs, err := p.ReleaseList(app, types.ReleaseListOptions{Count: 1})
	if err != nil {
		return nil, err
	}

	if len(rs) > 0 {
		r.Build = rs[0].Build
		r.Env = rs[0].Env
	}

	return r, nil
}
