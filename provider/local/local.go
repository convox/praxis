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
	"strings"
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
var Logger = logger.New("ns=provider.local")

type Provider struct {
	Frontend string
	Name     string
	Root     string
	Test     bool

	db *bolt.DB
}

// FromEnv returns a new local.Provider from env vars
func FromEnv() (*Provider, error) {
	p := &Provider{
		Frontend: coalesce(os.Getenv("PROVIDER_FRONTEND"), "10.42.84.0"),
		Name:     coalesce(os.Getenv("NAME"), "convox"),
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
	// if os.Getenv("PROVIDER_LOCAL_SKIP_FRONTEND_CHECK") != "true" {
	//   if err := p.checkFrontend(); err != nil {
	//     return err
	//   }
	// }

	if err := os.MkdirAll(p.Root, 0700); err != nil {
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

	go p.workers()

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

func (p *Provider) balancerRegister(app string, balancer manifest.Balancer) error {
	if p.Frontend == "none" {
		return nil
	}

	host := fmt.Sprintf("%s.%s.%s", balancer.Name, app, p.Name)

	for _, e := range balancer.Endpoints {
		data, err := exec.Command("docker", "inspect", "-f", "{{json .HostConfig.PortBindings}}", fmt.Sprintf("%s-%s-%s-%s", p.Name, app, balancer.Name, e.Port)).CombinedOutput()
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

func (p *Provider) balancerRunning(app string, balancer manifest.Balancer) bool {
	for _, e := range balancer.Endpoints {
		data, err := exec.Command("docker", "inspect", fmt.Sprintf("%s-%s-%s-%s", p.Name, app, balancer.Name, e.Port), "--format", "{{.State.Running}}").CombinedOutput()
		if err != nil {
			return false
		}

		if strings.HasPrefix(string(data), "false") {
			return false
		}
	}

	return true
}

func (p *Provider) balancerStart(app string, balancer manifest.Balancer) error {
	for _, e := range balancer.Endpoints {
		name := fmt.Sprintf("%s-%s-%s-%s", p.Name, app, balancer.Name, e.Port)

		exec.Command("docker", "rm", "-f", name).Run()

		command := []string{}

		switch {
		case e.Redirect != "":
			command = []string{"balancer", e.Protocol, "redirect", e.Redirect}
		case e.Target != "":
			command = []string{"balancer", e.Protocol, "target", e.Target}
		default:
			return fmt.Errorf("invalid balancer endpoint: %s:%s", balancer.Name, e.Port)
		}

		sys, err := p.SystemGet()
		if err != nil {
			return err
		}

		rp := rand.Intn(40000) + 20000

		args := []string{"run", "--detach", "--name", name}

		args = append(args, "--label", fmt.Sprintf("convox.app=%s", app))
		args = append(args, "--label", fmt.Sprintf("convox.balancer=%s", balancer.Name))
		args = append(args, "--label", fmt.Sprintf("convox.rack=%s", p.Name))
		args = append(args, "--label", "convox.type=balancer")

		args = append(args, "-e", fmt.Sprintf("APP=%s", app))
		args = append(args, "-e", fmt.Sprintf("RACK=%s", p.Name))

		hostname, err := os.Hostname()
		if err != nil {
			return err
		}

		args = append(args, "-e", fmt.Sprintf("RACK_URL=https://%s@%s:3000", os.Getenv("PASSWORD"), hostname))
		args = append(args, "--link", hostname)

		args = append(args, "-p", fmt.Sprintf("%d:3000", rp))

		args = append(args, sys.Image)
		args = append(args, command...)

		if err := exec.Command("docker", args...).Run(); err != nil {
			return err
		}

		if err := p.balancerRegister(app, balancer); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) serviceStart(m *manifest.Manifest, app, service, release string) error {
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

	k, err := types.Key(6)
	if err != nil {
		return err
	}

	_, err = p.ProcessStart(app, types.ProcessRunOptions{
		Command:     s.Command,
		Environment: senv,
		Name:        fmt.Sprintf("%s-%s-%s-%s", p.Name, app, service, k),
		Release:     release,
		Service:     service,
	})
	if err != nil {
		return err
	}

	return nil
}
