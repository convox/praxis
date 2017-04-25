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

	if len(opts.Env) > 0 {
		r.Env = opts.Env
	}

	if err := p.storageStore(fmt.Sprintf("apps/%s/releases/%s/release.json", app, r.Id), r); err != nil {
		return nil, err
	}

	if !p.Test {
		go func() {
			if err := p.releasePromote(app, r.Id); err != nil {
				r.Error = err.Error()
				r.Status = "failed"
			} else {
				r.Status = "complete"
			}

			p.storageStore(fmt.Sprintf("apps/%s/releases/%s/release.json", app, r.Id), r)
		}()
	}

	return r, nil
}

func (p *Provider) ReleaseGet(app, id string) (release *types.Release, err error) {
	err = p.storageLoad(fmt.Sprintf("apps/%s/releases/%s/release.json", app, id), &release)
	return
}

func (p *Provider) ReleaseList(app string) (types.Releases, error) {
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

		releases[i] = *release
	}

	sort.Slice(releases, func(i, j int) bool { return releases[j].Created.Before(releases[i].Created) })

	return releases, nil
}

func (p *Provider) ReleaseLogs(app, id string) (io.ReadCloser, error) {
	key := fmt.Sprintf("apps/%s/releases/%s/log", app, id)

	r, err := p.ReleaseGet(app, id)
	if err != nil {
		return nil, err
	}

	switch r.Status {
	case "complete", "failed":
	default:
		p.waitForRelease(app, id, nil)
	}

	lr, lw := io.Pipe()

	go func() {
		defer lw.Close()
		p.storageLogRead(key, func(at time.Time, entry []byte) {
			lw.Write(entry)
		})
	}()
	if err != nil {
		return nil, err
	}

	return lr, nil
}

func (p *Provider) releaseFork(app string) (*types.Release, error) {
	r := &types.Release{
		Id:      types.Id("R", 10),
		App:     app,
		Status:  "created",
		Created: time.Now().UTC(),
	}

	rs, err := p.ReleaseList(app)
	if err != nil {
		return nil, err
	}

	if len(rs) > 0 {
		r.Build = rs[0].Build
		r.Env = rs[0].Env
	}

	return r, nil
}

func (p *Provider) releasePromote(app, release string) error {
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

	r.Status = "promoting"

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

	m, err := manifest.Load([]byte(b.Manifest))
	if err != nil {
		return err
	}

	if err := p.converge(app); err != nil {
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

	r.Status = "complete"

	if err := p.storageStore(fmt.Sprintf("apps/%s/releases/%s/release.json", app, release), r); err != nil {
		return err
	}

	return nil
}

func (p *Provider) waitForRelease(app, id string, fn func()) {
	for {
		time.Sleep(1 * time.Second)

		r, err := p.ReleaseGet(app, id)
		if err != nil {
			continue
		}

		if r.Status == "complete" || r.Status == "failed" {
			break
		}
	}

	if fn != nil {
		fn()
	}
}
