package local

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"sort"
	"strings"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

func (p *Provider) ResourceGet(app, name string) (*types.Resource, error) {
	rs, err := p.ResourceList(app)
	if err != nil {
		return nil, err
	}

	for _, r := range rs {
		if r.Name == name {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("resource not found: %s", name)
}

func (p *Provider) ResourceList(app string) (types.Resources, error) {
	log := p.logger("ResourceList").Append("app=%q", app)

	m, _, err := helpers.AppManifest(p, app)
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	rs := make(types.Resources, len(m.Resources))

	for i, r := range m.Resources {
		e, err := resourceURL(app, r.Type, r.Name)
		if err != nil {
			return nil, errors.WithStack(log.Error(err))
		}

		rs[i] = types.Resource{
			Name:     r.Name,
			Endpoint: e,
			Type:     r.Type,
		}
	}

	sort.Slice(rs, func(i, j int) bool { return rs[i].Name < rs[j].Name })

	return rs, log.Success()
}

func (p *Provider) ResourceProxy(app, resource string, in io.Reader) (io.ReadCloser, error) {
	_, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	r, err := p.ResourceGet(app, resource)
	if err != nil {
		return nil, err
	}

	port, err := resourcePort(r.Type)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%s.%s.resource.%s", p.Name, app, resource)

	data, err := exec.Command("docker", "inspect", name, "--format", "{{.NetworkSettings.IPAddress}}").CombinedOutput()
	if err != nil {
		return nil, err
	}

	ip := strings.TrimSpace(string(data))

	cn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err
	}

	go helpers.Stream(cn, in)

	return cn, nil
}
