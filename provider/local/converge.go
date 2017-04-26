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

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func (p *Provider) converge(app string) error {
	log := Logger.At("converge").Start()

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

	if err := p.balancersConverge(app, r.Id, m.Balancers); err != nil {
		log.Error(err)
	}

	if err := p.endpointsConverge(app, r.Id, m.Services); err != nil {
		log.Error(err)
	}

	if err := p.servicesConverge(app, r.Id, m.Services); err != nil {
		log.Error(err)
	}

	log.Success()
	return nil
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
	args := []string{"ps", "--format", "{{.ID}}"}

	for k, v := range labels {
		args = append(args, "--filter", fmt.Sprintf("label=%s=%s", k, v))
	}

	data, err := exec.Command("docker", args...).CombinedOutput()
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

func (p *Provider) containersKillOutdated(kind, app, release string) error {
	acs, err := containersByLabels(map[string]string{
		"convox.type": kind,
		"convox.rack": p.Name,
		"convox.app":  app,
	})
	if err != nil {
		return err
	}

	cs, err := containersByLabels(map[string]string{
		"convox.type":    kind,
		"convox.rack":    p.Name,
		"convox.app":     app,
		"convox.release": release,
	})
	if err != nil {
		return err
	}

	for _, id := range diff(acs, cs) {
		if err := exec.Command("docker", "kill", id).Run(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) balancersConverge(app, release string, balancers manifest.Balancers) error {
	if err := p.containersKillOutdated("balancer", app, release); err != nil {
		return err
	}

	for _, b := range balancers {
		if !p.balancerRunning(app, release, b.Name) {
			if err := p.balancerStart(app, release, b); err != nil {
				return err
			}
		}

		if err := p.balancerRegister(app, b); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) balancerRegister(app string, balancer manifest.Balancer) error {
	if p.Frontend == "none" {
		return nil
	}

	host := fmt.Sprintf("%s-%s.%s", app, balancer.Name, p.Name)

	for _, e := range balancer.Endpoints {
		name := fmt.Sprintf("%s.%s.balancer.%s.%s", p.Name, app, balancer.Name, e.Port)

		port, err := containerBinding(name, "3000/tcp")
		if err != nil {
			return err
		}
		if port == "" {
			return fmt.Errorf("balancer not bound to 3000/tcp: %s", name)
		}

		uv := url.Values{}
		uv.Add("port", e.Port)
		uv.Add("target", fmt.Sprintf("localhost:%s", port))

		req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:9477/endpoints/%s", p.Frontend, host), bytes.NewReader([]byte(uv.Encode())))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		defer res.Body.Close()
	}

	return nil
}

func (p *Provider) balancerRunning(app, release, balancer string) bool {
	cs, err := containersByLabels(map[string]string{
		"convox.type":     "balancer",
		"convox.rack":     p.Name,
		"convox.app":      app,
		"convox.release":  release,
		"convox.balancer": balancer,
	})
	if err != nil {
		return false
	}
	if len(cs) == 0 {
		return false
	}

	return true
}

func (p *Provider) balancerStart(app, release string, balancer manifest.Balancer) error {
	for _, e := range balancer.Endpoints {
		name := fmt.Sprintf("%s.%s.balancer.%s.%s", p.Name, app, balancer.Name, e.Port)

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

		hostname, err := os.Hostname()
		if err != nil {
			return err
		}

		args := []string{"run", "--rm", "--detach"}

		args = append(args, "--name", name)
		args = append(args, "--label", fmt.Sprintf("convox.app=%s", app))
		args = append(args, "--label", fmt.Sprintf("convox.balancer=%s", balancer.Name))
		args = append(args, "--label", fmt.Sprintf("convox.rack=%s", p.Name))
		args = append(args, "--label", fmt.Sprintf("convox.release=%s", release))
		args = append(args, "--label", "convox.type=balancer")
		args = append(args, "-e", fmt.Sprintf("APP=%s", app))
		args = append(args, "-e", fmt.Sprintf("RACK=%s", p.Name))
		args = append(args, "-e", fmt.Sprintf("RACK_URL=https://%s:3000", hostname))
		args = append(args, "--link", hostname)
		args = append(args, "-p", fmt.Sprintf("%d:3000", rp))
		args = append(args, sys.Image)
		args = append(args, command...)

		if err := exec.Command("docker", args...).Run(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) endpointsConverge(app, release string, services manifest.Services) error {
	if err := p.containersKillOutdated("endpoint", app, release); err != nil {
		return err
	}

	for _, s := range services {
		if s.Port.Port == 0 {
			continue
		}

		if !p.endpointRunning(app, release, s.Name) {
			if err := p.endpointStart(app, release, s); err != nil {
				return err
			}
		}

		if err := p.endpointRegister(app, s); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) endpointRegister(app string, s manifest.Service) error {
	if p.Frontend == "none" {
		return nil
	}

	if s.Port.Port == 0 {
		return fmt.Errorf("no endpoint")
	}

	host := fmt.Sprintf("%s-%s.%s", app, s.Name, p.Name)
	name := fmt.Sprintf("%s.%s.endpoint.%s.%d", p.Name, app, s.Name, s.Port.Port)

	port, err := containerBinding(name, "3000/tcp")
	if err != nil {
		return err
	}
	if port == "" {
		return fmt.Errorf("balancer not bound to 3000/tcp: %s", name)
	}

	uv := url.Values{}
	uv.Add("port", "443")
	uv.Add("target", fmt.Sprintf("localhost:%s", port))

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:9477/endpoints/%s", p.Frontend, host), bytes.NewReader([]byte(uv.Encode())))
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

func (p *Provider) endpointRunning(app, release, service string) bool {
	cs, err := containersByLabels(map[string]string{
		"convox.type":    "endpoint",
		"convox.rack":    p.Name,
		"convox.app":     app,
		"convox.release": release,
		"convox.service": service,
	})
	if err != nil {
		return false
	}
	if len(cs) == 0 {
		return false
	}

	return true
}

func (p *Provider) endpointStart(app, release string, s manifest.Service) error {
	if s.Port.Port == 0 {
		return fmt.Errorf("no endpoint")
	}

	name := fmt.Sprintf("%s.%s.endpoint.%s.%d", p.Name, app, s.Name, s.Port.Port)

	exec.Command("docker", "rm", "-f", name).Run()

	command := []string{"balancer", "https", "target", fmt.Sprintf("%s://%s:%d", s.Port.Scheme, s.Name, s.Port.Port)}

	sys, err := p.SystemGet()
	if err != nil {
		return err
	}

	rp := rand.Intn(40000) + 20000

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	args := []string{"run", "--rm", "--detach"}

	args = append(args, "--name", name)
	args = append(args, "--label", fmt.Sprintf("convox.app=%s", app))
	args = append(args, "--label", fmt.Sprintf("convox.rack=%s", p.Name))
	args = append(args, "--label", fmt.Sprintf("convox.release=%s", release))
	args = append(args, "--label", fmt.Sprintf("convox.service=%s", s.Name))
	args = append(args, "--label", "convox.type=endpoint")
	args = append(args, "-e", fmt.Sprintf("APP=%s", app))
	args = append(args, "-e", fmt.Sprintf("RACK=%s", p.Name))
	args = append(args, "-e", fmt.Sprintf("RACK_URL=https://%s:3000", hostname))
	args = append(args, "--link", hostname)
	args = append(args, "-p", fmt.Sprintf("%d:3000", rp))
	args = append(args, sys.Image)
	args = append(args, command...)

	if err := exec.Command("docker", args...).Run(); err != nil {
		return err
	}

	return nil
}

func (p *Provider) servicesConverge(app, release string, services manifest.Services) error {
	if err := p.containersKillOutdated("service", app, release); err != nil {
		return err
	}

	for _, s := range services {
		if !p.serviceRunning(app, release, s.Name) {
			if err := p.serviceStart(app, release, s); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Provider) serviceRunning(app, release, service string) bool {
	cs, err := containersByLabels(map[string]string{
		"convox.type":    "service",
		"convox.rack":    p.Name,
		"convox.app":     app,
		"convox.release": release,
		"convox.service": service,
	})
	if err != nil {
		return false
	}
	if len(cs) == 0 {
		return false
	}

	return true
}

func (p *Provider) serviceStart(app, release string, service manifest.Service) error {
	r, err := p.ReleaseGet(app, release)
	if err != nil {
		return err
	}

	senv, err := service.Env(r.Env)
	if err != nil {
		return err
	}

	k, err := types.Key(6)
	if err != nil {
		return err
	}

	_, err = p.ProcessStart(app, types.ProcessRunOptions{
		Command:     service.Command,
		Environment: senv,
		Name:        fmt.Sprintf("%s.%s.service.%s.%s", p.Name, app, service.Name, k),
		Release:     release,
		Service:     service.Name,
		Type:        "service",
	})
	if err != nil {
		return err
	}

	return nil
}
