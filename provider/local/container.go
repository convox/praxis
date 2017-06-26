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
	"strconv"
	"strings"
	"sync"
)

type container struct {
	Command  []string
	Env      map[string]string
	Hostname string
	Image    string
	Labels   map[string]string
	Id       string
	Memory   int
	Name     string
	Port     containerPort
	Volumes  []string
}

type containerPort struct {
	Container int
	Host      int
}

// func (p *Provider) containerConverge(c container, app, release string) (string, error) {
//   args := []string{}

//   for k, v := range c.Labels {
//     args = append(args, "--filter", fmt.Sprintf("label=%s=%s", k, v))
//   }

//   ids, err := containerList(args...)
//   if err != nil {
//     return "", errors.WithStack(err)
//   }

//   id := ""

//   switch len(ids) {
//   case 0:
//     p.storageLogWrite(fmt.Sprintf("apps/%s/releases/%s/log", app, release), []byte(fmt.Sprintf("starting: %s\n", c.Name)))

//     i, err := p.containerStart(c, app, release)
//     if err != nil {
//       return "", errors.WithStack(err)
//     }

//     id = i
//   case 1:
//     id = ids[0]
//   default:
//     return "", fmt.Errorf("matched more than one container")
//   }

//   return id, nil
// }

func (p *Provider) containerRegister(c container) error {
	if p.Frontend == "none" || c.Hostname == "" || c.Port.Container == 0 || c.Port.Host == 0 {
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

	if m := c.Memory; m > 0 {
		args = append(args, "--memory-reservation", fmt.Sprintf("%dm", m))
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
	args = append(args, "-e", fmt.Sprintf("RACK_URL=https://%s:3000", hostname))
	args = append(args, "-e", fmt.Sprintf("RELEASE=%s", release))
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

func (p *Provider) containerStop(id string) error {
	return exec.Command("docker", "stop", "--time", "3", id).Run()
}

func (p *Provider) containerStopAsync(id string, wg *sync.WaitGroup) {
	defer wg.Done()
	p.containerStop(id)
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

func containersByLabels(labels map[string]string) ([]container, error) {
	args := []string{}

	for k, v := range labels {
		args = append(args, "--filter", fmt.Sprintf("label=%s=%s", k, v))
	}

	return containerList(args...)
}

func containerIDs(args ...string) ([]string, error) {
	as := []string{"ps", "--format", "{{.ID}}"}

	as = append(as, args...)

	data, err := exec.Command("docker", as...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	all := strings.TrimSpace(string(data))

	if all == "" {
		return []string{}, nil
	}

	return strings.Split(all, "\n"), nil
}

func containerList(args ...string) ([]container, error) {
	ids, err := containerIDs(args...)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []container{}, nil
	}

	as := []string{"inspect", "--format", "{{json .}}"}
	as = append(as, ids...)

	data, err := exec.Command("docker", as...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	cs := []container{}

	d := json.NewDecoder(bytes.NewReader(data))

	for d.More() {
		var c struct {
			Id     string
			Config struct {
				Labels map[string]string
			}
			HostConfig struct {
				PortBindings map[string][]struct {
					HostIp   string
					HostPort string
				}
			}
			Name string
		}

		if err := d.Decode(&c); err != nil {
			return nil, err
		}

		cc := container{
			Id:     c.Id,
			Labels: c.Config.Labels,
			Name:   c.Name[1:],
			Port: containerPort{
				Container: 0,
				Host:      0,
			},
		}

		if len(c.HostConfig.PortBindings) == 1 {
			for k, v := range c.HostConfig.PortBindings {
				if len(v) != 1 {
					continue
				}

				cps := strings.Split(k, "/")[0]

				cpi, err := strconv.Atoi(cps)
				if err != nil {
					return nil, err
				}

				hpi, err := strconv.Atoi(v[0].HostPort)
				if err != nil {
					return nil, err
				}

				cc.Port = containerPort{
					Container: cpi,
					Host:      hpi,
				}
			}
		}

		cs = append(cs, cc)
	}

	return cs, nil
}
