package local

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/convox/praxis/manifest"
	shellquote "github.com/kballard/go-shellquote"
)

type container struct {
	Command  []string
	Env      map[string]string
	Hostname string
	Image    string
	Labels   map[string]string
	Id       string
	Name     string
	Port     containerPort
	Volumes  []string
}

type containerPort struct {
	Container int
	Host      int
}

func (p *Provider) converge(app string) error {
	log := Logger.At("converge").Namespace("app=%s", app).Start()

	a, err := p.AppGet(app)
	if err != nil {
		log.Error(err)
		return err
	}

	if a.Release == "" {
		return nil
	}

	r, err := p.ReleaseGet(a.Name, a.Release)
	if err != nil {
		log.Error(err)
		return err
	}

	b, err := p.BuildGet(a.Name, r.Build)
	if err != nil {
		log.Error(err)
		return err
	}

	m, err := manifest.Load([]byte(b.Manifest))
	if err != nil {
		log.Error(err)
		return err
	}

	cs := []container{}

	c, err := p.balancerContainers(m.Balancers, app, r.Id)
	if err != nil {
		return err
	}

	cs = append(cs, c...)

	c, err = p.resourceContainers(m.Resources, app, r.Id)
	if err != nil {
		return err
	}

	cs = append(cs, c...)

	c, err = p.serviceContainers(m.Services, app, r.Id)
	if err != nil {
		return err
	}

	cs = append(cs, c...)

	for i, c := range cs {
		id, err := p.containerConverge(c, app, r.Id)
		if err != nil {
			return err
		}

		cs[i].Id = id

		if err := p.containerRegister(cs[i]); err != nil {
			return err
		}
	}

	running, err := containersByLabels(map[string]string{
		"convox.rack": p.Name,
		"convox.app":  app,
	})
	if err != nil {
		return err
	}

	ps, err := containersByLabels(map[string]string{
		"convox.rack": p.Name,
		"convox.app":  app,
		"convox.type": "process",
	})
	if err != nil {
		return err
	}

	for _, rc := range running {
		found := false

		for _, c := range cs {
			if c.Id == rc {
				found = true
				break
			}
		}

		// dont stop oneoff processes
		for _, pc := range ps {
			if rc == pc {
				found = true
				break
			}
		}

		if !found {
			log.Successf("action=kill id=%s", rc)
			exec.Command("docker", "kill", rc).Run()
		}
	}

	log.Success()
	return nil
}

func (p *Provider) containerConverge(c container, app, release string) (string, error) {
	log := Logger.At("converge.container").Namespace("app=%s name=%s", app, c.Name).Start()

	args := []string{}

	for k, v := range c.Labels {
		args = append(args, "--filter", fmt.Sprintf("label=%s=%s", k, v))
	}

	ids, err := containerList(args...)
	if err != nil {
		return "", err
	}

	id := ""

	switch len(ids) {
	case 0:
		i, err := p.containerStart(c, app, release)
		if err != nil {
			return "", err
		}

		id = i

		log = log.Namespace("action=start")
	case 1:
		id = ids[0]

		log = log.Namespace("action=found")
	default:
		return "", fmt.Errorf("matched more than one container")
	}

	log.Success()
	return id, nil
}

func (p *Provider) containerRegister(c container) error {
	if c.Hostname == "" || c.Port.Container == 0 || c.Port.Host == 0 {
		return nil
	}

	bind, err := containerBinding(c.Id, fmt.Sprintf("%d/tcp", c.Port.Container))
	if err != nil {
		return err
	}

	uv := url.Values{}
	uv.Add("port", strconv.Itoa(c.Port.Host))
	uv.Add("target", fmt.Sprintf("localhost:%s", bind))

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:9477/endpoints/%s", p.Frontend, c.Hostname), bytes.NewReader([]byte(uv.Encode())))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return nil
}

func (p *Provider) containerStart(c container, app, release string) (string, error) {
	if c.Name == "" {
		return "", fmt.Errorf("name required")
	}

	args := []string{"run", "--detach"}

	args = append(args, "--name", c.Name)

	for k, v := range c.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, v))
	}

	for k, v := range c.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	if p := c.Port.Container; p != 0 {
		args = append(args, "-p", fmt.Sprintf("%d:%d", rand.Intn(40000)+20000, p))
	}

	for _, v := range c.Volumes {
		args = append(args, "-v", v)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	args = append(args, "-e", fmt.Sprintf("APP=%s", app))
	args = append(args, "-e", fmt.Sprintf("RACK=%s", p.Name))
	args = append(args, "-e", fmt.Sprintf("RACK_URL=https://%s:3000", hostname))
	args = append(args, "--link", hostname)

	args = append(args, c.Image)
	args = append(args, c.Command...)

	exec.Command("docker", "rm", "-f", c.Name).Run()

	data, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return "", err
	}

	id := strings.TrimSpace(string(data))

	if len(id) < 12 {
		return "", fmt.Errorf("unable to start container")
	}

	return id[0:12], nil
}

func containerBinding(id string, bind string) (string, error) {
	data, err := exec.Command("docker", "inspect", "-f", "{{json .HostConfig.PortBindings}}", id).CombinedOutput()
	if err != nil {
		return "", err
	}

	var bindings map[string][]struct {
		HostPort string
	}

	if err := json.Unmarshal(data, &bindings); err != nil {
		return "", err
	}

	b, ok := bindings[bind]
	if !ok {
		return "", nil
	}
	if len(b) < 1 {
		return "", nil
	}

	return b[0].HostPort, nil
}

func containersByLabels(labels map[string]string) ([]string, error) {
	args := []string{}

	for k, v := range labels {
		args = append(args, "--filter", fmt.Sprintf("label=%s=%s", k, v))
	}

	return containerList(args...)
}

func containerList(args ...string) ([]string, error) {
	as := []string{"ps", "--format", "{{.ID}}"}
	as = append(as, args...)

	data, err := exec.Command("docker", as...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	cs := []string{}

	s := bufio.NewScanner(bytes.NewReader(data))

	for s.Scan() {
		cs = append(cs, s.Text())
	}

	return cs, nil
}

func resourcePort(kind string) (int, error) {
	switch kind {
	case "postgres":
		return 5432, nil
	}

	return 0, fmt.Errorf("unknown resource type: %s", kind)
}

func resourceURL(app, kind, name string) (string, error) {
	switch kind {
	case "postgres":
		return fmt.Sprintf("postgres://postgres:password@%s.resource.%s.convox:5432/app?sslmode=disable", name, app), nil
	}

	return "", fmt.Errorf("unknown resource type: %s", kind)
}

func resourceVolumes(app, kind, name string) ([]string, error) {
	switch kind {
	case "postgres":
		return []string{fmt.Sprintf("/var/convox/%s/resource/%s:/var/lib/postgresql/data", app, name)}, nil
	}

	return []string{}, fmt.Errorf("unknown resource type: %s", kind)
}

func (p *Provider) balancerContainers(balancers manifest.Balancers, app, release string) ([]container, error) {
	cs := []container{}

	sys, err := p.SystemGet()
	if err != nil {
		return nil, err
	}

	for _, b := range balancers {
		for _, e := range b.Endpoints {
			command := []string{}

			switch {
			case e.Redirect != "":
				command = []string{"balancer", e.Protocol, "redirect", e.Redirect}
			case e.Target != "":
				command = []string{"balancer", e.Protocol, "target", e.Target}
			default:
				return nil, fmt.Errorf("invalid balancer endpoint: %s:%s", b.Name, e.Port)
			}

			cs = append(cs, container{
				Name:     fmt.Sprintf("%s.%s.balancer.%s", p.Name, app, b.Name),
				Hostname: fmt.Sprintf("%s.balancer.%s.%s", b.Name, app, p.Name),
				Port: containerPort{
					Host:      443,
					Container: 3000,
				},
				Image:   sys.Image,
				Command: command,
				Labels: map[string]string{
					"convox.rack":    p.Name,
					"convox.app":     app,
					"convox.release": release,
					"convox.type":    "balancer",
					"convox.name":    b.Name,
					"convox.port":    e.Port,
				},
			})
		}
	}

	return cs, nil
}

func (p *Provider) resourceContainers(resources manifest.Resources, app, release string) ([]container, error) {
	cs := []container{}

	for _, r := range resources {
		rp, err := resourcePort(r.Type)
		if err != nil {
			return nil, err
		}

		vs, err := resourceVolumes(app, r.Type, r.Name)
		if err != nil {
			return nil, err
		}

		cs = append(cs, container{
			Name:     fmt.Sprintf("%s.%s.resource.%s", p.Name, app, r.Name),
			Hostname: fmt.Sprintf("%s.resource.%s.%s", r.Name, app, p.Name),
			Port: containerPort{
				Host:      rp,
				Container: rp,
			},
			Image:   fmt.Sprintf("convox/%s", r.Type),
			Volumes: vs,
			Labels: map[string]string{
				"convox.rack":     p.Name,
				"convox.app":      app,
				"convox.release":  release,
				"convox.type":     "resource",
				"convox.name":     r.Name,
				"convox.resource": r.Type,
			},
		})
	}

	return cs, nil
}

func (p *Provider) serviceContainers(services manifest.Services, app, release string) ([]container, error) {
	cs := []container{}

	sys, err := p.SystemGet()
	if err != nil {
		return nil, err
	}

	r, err := p.ReleaseGet(app, release)
	if err != nil {
		return nil, err
	}

	b, err := p.BuildGet(app, r.Build)
	if err != nil {
		return nil, err
	}

	m, err := manifest.Load([]byte(b.Manifest))
	if err != nil {
		return nil, err
	}

	for _, s := range services {
		if s.Port.Port > 0 {
			cs = append(cs, container{
				Name:     fmt.Sprintf("%s.%s.endpoint.%s", p.Name, app, s.Name),
				Hostname: fmt.Sprintf("%s.service.%s.%s", s.Name, app, p.Name),
				Port: containerPort{
					Host:      443,
					Container: 3000,
				},
				Image:   sys.Image,
				Command: []string{"balancer", "https", "target", fmt.Sprintf("%s://%s:%d", s.Port.Scheme, s.Name, s.Port.Port)},
				Labels: map[string]string{
					"convox.rack":    p.Name,
					"convox.app":     app,
					"convox.release": release,
					"convox.type":    "endpoint",
					"convox.name":    s.Name,
					"convox.service": s.Name,
					"convox.port":    strconv.Itoa(s.Port.Port),
				},
			})
		}

		cmd, err := shellquote.Split(s.Command.Production)
		if err != nil {
			return nil, err
		}

		env, err := s.Env(r.Env)
		if err != nil {
			return nil, err
		}

		// copy the map so we can hold on to it
		e := map[string]string{}

		for k, v := range env {
			e[k] = v
		}

		// add resources
		for _, sr := range s.Resources {
			for _, r := range m.Resources {
				if r.Name == sr {
					u, err := resourceURL(app, r.Type, r.Name)
					if err != nil {
						return nil, err
					}

					e[fmt.Sprintf("%s_URL", strings.ToUpper(sr))] = u
				}
			}
		}

		cs = append(cs, container{
			Name:    fmt.Sprintf("%s.%s.service.%s.1", p.Name, app, s.Name),
			Image:   fmt.Sprintf("%s/%s/%s:%s", p.Name, app, s.Name, r.Build),
			Command: cmd,
			Env:     e,
			Volumes: s.Volumes,
			Labels: map[string]string{
				"convox.rack":    p.Name,
				"convox.app":     app,
				"convox.release": release,
				"convox.type":    "service",
				"convox.name":    s.Name,
				"convox.service": s.Name,
				"convox.index":   "1",
			},
		})
	}

	return cs, nil
}
