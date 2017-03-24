package local

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
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
	err = p.storageLoad(fmt.Sprintf("/apps/%s/releases/%s/release.json", app, id), &release)
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
		return p.storageRead(key)
	default:
		r, w := io.Pipe()

		log, err := p.storageTail(fmt.Sprintf("apps/%s/releases/%s/log", app, id))
		if err != nil {
			return nil, err
		}

		go io.Copy(w, log)

		go p.waitForRelease(app, id, func() {
			w.Close()
		})

		return r, nil
	}
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

	log, err := p.storageWrite(fmt.Sprintf("apps/%s/releases/%s/log", app, release))
	if err != nil {
		return err
	}

	defer log.Close()

	for _, s := range m.Services {
		fmt.Fprintf(log, "starting service: %s\n", s.Name)

		if err := p.startService(m, app, s.Name, r.Id); err != nil {
			return err
		}
	}

	for _, b := range m.Balancers {
		fmt.Fprintf(log, "starting balancer: %s\n", b.Name)

		if err := p.startBalancer(app, b); err != nil {
			return err
		}
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

func (p *Provider) startBalancer(app string, balancer manifest.Balancer) error {
	for _, e := range balancer.Endpoints {
		name := fmt.Sprintf("balancer-%s-%s-%s", app, balancer.Name, e.Port)

		command := ""

		switch {
		case e.Redirect != "":
			command = fmt.Sprintf("proxy %s redirect %s", e.Protocol, e.Redirect)
		case e.Target != "":
			command = fmt.Sprintf("proxy %s target %s", e.Protocol, e.Target)
		default:
			return fmt.Errorf("invalid balancer endpoint: %s:%s", balancer.Name, e.Port)
		}

		port, err := strconv.Atoi(e.Port)
		if err != nil {
			return err
		}

		sys, err := p.SystemGet()
		if err != nil {
			return err
		}

		rp := rand.Intn(40000) + 20000

		fmt.Printf("rp = %+v\n", rp)

		opts := types.ProcessRunOptions{
			Command: command,
			Image:   sys.Image,
			Name:    name,
			Ports:   map[int]int{rp: 3000},
			Stream:  types.Stream{Writer: os.Stdout},
		}

		if _, err := p.ProcessStart(app, opts); err != nil {
			return err
		}

		uv := url.Values{}
		uv.Add("port", strconv.Itoa(port))
		uv.Add("target", fmt.Sprintf("localhost:%d", rp))

		fmt.Printf("uv = %+v\n", uv)

		host := fmt.Sprintf("%s.%s.convox", balancer.Name, app)

		fmt.Printf("host = %+v\n", host)

		req, err := http.NewRequest("POST", fmt.Sprintf("http://10.42.84.0:9477/endpoints/%s", host), bytes.NewReader([]byte(uv.Encode())))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		fmt.Printf("res = %+v\n", res)
	}

	return nil
}

func (p *Provider) startService(m *manifest.Manifest, app, service, release string) error {
	pss, err := p.ProcessList(app, types.ProcessListOptions{Service: service})
	if err != nil {
		return err
	}

	for _, ps := range pss {
		if err := p.ProcessStop(app, ps.Id); err != nil {
			return err
		}
	}

	s, err := m.Services.Find(service)
	if err != nil {
		return err
	}

	r, err := p.ReleaseGet(app, release)
	if err != nil {
		return err
	}

	senv, err := s.Env(r.Env)
	if err != nil {
		return err
	}

	_, err = p.ProcessStart(app, types.ProcessRunOptions{
		Command:     s.Command,
		Environment: senv,
		Release:     release,
		Service:     service,
	})
	if err != nil {
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

	fn()
}
