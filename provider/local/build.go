package local

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/convox/praxis/types"
)

func (p *Provider) BuildCreate(app, url string, opts types.BuildCreateOptions) (*types.Build, error) {
	_, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	bid := types.Id("B", 10)

	build := &types.Build{
		Id:     bid,
		App:    app,
		Status: "created",
	}

	if err := p.Store(fmt.Sprintf("apps/%s/builds/%s", app, bid), build); err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	args := []string{"run"}
	args = append(args, "--detach", "-i")
	args = append(args, "--link", hostname, "-e", fmt.Sprintf("RACK_URL=https://%s:3000", hostname))
	args = append(args, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	args = append(args, "-e", fmt.Sprintf("BUILD_APP=%s", app))
	args = append(args, "convox/praxis", "build", "-id", bid, "-url", url)

	cmd := exec.Command("docker", args...)

	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	if _, err := p.BuildUpdate(app, bid, types.BuildUpdateOptions{Process: strings.TrimSpace(string(data))}); err != nil {
		return nil, err
	}

	return build, nil
}

func (p *Provider) BuildGet(app, id string) (build *types.Build, err error) {
	err = p.Load(fmt.Sprintf("apps/%s/builds/%s", app, id), &build)
	return
}

func (p *Provider) BuildLogs(app, id string) (io.Reader, error) {
	build, err := p.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	return p.Logs(build.Process)
}

func (p *Provider) BuildUpdate(app, id string, opts types.BuildUpdateOptions) (*types.Build, error) {
	build, err := p.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	if opts.Manifest != "" {
		build.Manifest = opts.Manifest
	}

	if opts.Process != "" {
		build.Process = opts.Process
	}

	if opts.Release != "" {
		build.Release = opts.Release
	}

	if opts.Status != "" {
		build.Status = opts.Status
	}

	if err := p.Store(fmt.Sprintf("apps/%s/builds/%s", app, id), build); err != nil {
		return nil, err
	}

	return build, nil
}
