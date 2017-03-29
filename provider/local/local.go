package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/convox/logger"
	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

var (
	customTopic       = os.Getenv("CUSTOM_TOPIC")
	notificationTopic = os.Getenv("NOTIFICATION_TOPIC")
	sortableTime      = "20060102.150405.000000000"
)

// Logger is a package-wide logger
var Logger = logger.New("ns=provider.aws")

type Provider struct {
	Frontend string
	Root     string
	Test     bool

	db *bolt.DB
}

// FromEnv returns a new local.Provider from env vars
func FromEnv() (*Provider, error) {
	p := &Provider{
		Frontend: coalesce(os.Getenv("PROVIDER_FRONTEND"), "10.42.84.0"),
		Root:     coalesce(os.Getenv("PROVIDER_ROOT"), "/var/convox"),
	}

	if err := p.Init(); err != nil {
		return nil, err
	}

	return p, nil
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func (p *Provider) Init() error {
	if err := p.checkFrontend(); err != nil {
		return err
	}

	db, err := bolt.Open(filepath.Join(p.Root, "rack.db"), 0600, nil)
	if err != nil {
		return err
	}

	p.db = db

	if _, err := p.createRootBucket("rack"); err != nil {
		return err
	}

	return nil
}

func (p *Provider) createRootBucket(name string) (*bolt.Bucket, error) {
	tx, err := p.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket, err := tx.CreateBucketIfNotExists([]byte("rack"))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return bucket, err
}

func (p *Provider) checkFrontend() error {
	if p.Frontend == "none" {
		return nil
	}

	c := http.DefaultClient

	c.Timeout = 2 * time.Second

	if _, err := c.Get(fmt.Sprintf("http://%s:9477/endpoints", p.Frontend)); err != nil {
		return fmt.Errorf("unable to register with frontend")
	}

	return nil
}

func (p *Provider) registerBalancers() {
	tick := time.Tick(1 * time.Minute)

	if err := p.registerBalancersTick(); err != nil {
		fmt.Printf("err = %+v\n", err)
	}

	for range tick {
		if err := p.registerBalancersTick(); err != nil {
			fmt.Printf("err = %+v\n", err)
		}
	}
}

func (p *Provider) registerBalancersTick() error {
	if p.Frontend == "none" {
		return nil
	}

	apps, err := p.AppList()
	if err != nil {
		return err
	}

	for _, app := range apps {
		if app.Release == "" {
			continue
		}

		r, err := p.ReleaseGet(app.Name, app.Release)
		if err != nil {
			return err
		}

		b, err := p.BuildGet(app.Name, r.Build)
		if err != nil {
			return err
		}

		m, err := manifest.Load([]byte(b.Manifest))
		if err != nil {
			return err
		}

		for _, b := range m.Balancers {
			if err := p.registerBalancerWithFrontend(app.Name, b); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Provider) registerBalancerWithFrontend(app string, balancer manifest.Balancer) error {
	if p.Frontend == "none" {
		return nil
	}

	host := fmt.Sprintf("%s.%s.convox", balancer.Name, app)

	for _, e := range balancer.Endpoints {
		data, err := exec.Command("docker", "inspect", "-f", "{{json .HostConfig.PortBindings}}", fmt.Sprintf("balancer-%s-%s-%s", app, balancer.Name, e.Port)).CombinedOutput()
		if err != nil {
			continue
		}

		var bindings map[string][]struct {
			HostPort string
		}

		if err := json.Unmarshal(data, &bindings); err != nil {
			return err
		}

		bind, ok := bindings["3000/tcp"]
		if !ok || len(bind) < 1 || bind[0].HostPort == "" {
			return fmt.Errorf("invalid balancer binding")
		}

		port := bind[0].HostPort

		uv := url.Values{}
		uv.Add("port", e.Port)
		uv.Add("target", fmt.Sprintf("localhost:%s", port))

		req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:9477/endpoints/%s", p.Frontend, host), bytes.NewReader([]byte(uv.Encode())))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		if _, err := http.DefaultClient.Do(req); err != nil {
			return err
		}
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

		sys, err := p.SystemGet()
		if err != nil {
			return err
		}

		rp := rand.Intn(40000) + 20000

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

		if err := p.registerBalancerWithFrontend(app, balancer); err != nil {
			return err
		}

		go p.registerBalancers()
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
