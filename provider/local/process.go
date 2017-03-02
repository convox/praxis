package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func (p *Provider) ProcessList(app string, opts types.ProcessListOptions) (types.Processes, error) {
	args := []string{"ps"}

	if opts.Service != "" {
		args = append(args, "--filter", fmt.Sprintf("label=convox.service=%s", opts.Service))
	}

	args = append(args, "--filter", fmt.Sprintf("label=convox.app=%s", app))
	args = append(args, "--format", "{{json .}}")

	data, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	ps := types.Processes{}

	jd := json.NewDecoder(bytes.NewReader(data))

	for jd.More() {
		var dps struct {
			Command string
			ID      string
			Labels  string
		}

		if err := jd.Decode(&dps); err != nil {
			return nil, err
		}

		labels := map[string]string{}

		for _, kv := range strings.Split(dps.Labels, ",") {
			parts := strings.SplitN(kv, "=", 2)

			if len(parts) == 2 {
				labels[parts[0]] = parts[1]
			}
		}

		ps = append(ps, types.Process{
			Id:      dps.ID,
			App:     labels["convox.app"],
			Command: strings.Trim(dps.Command, `"`),
			Release: labels["convox.release"],
			Service: labels["convox.service"],
		})
	}

	return ps, nil
}

func (p *Provider) ProcessRun(app string, opts types.ProcessRunOptions) (int, error) {
	release, err := p.ReleaseGet(app, opts.Release)
	if err != nil {
		return 0, err
	}

	build, err := p.BuildGet(app, release.Build)
	if err != nil {
		return 0, err
	}

	m, err := manifest.Load([]byte(build.Manifest))
	if err != nil {
		return 0, err
	}

	service, err := m.Services.Find(opts.Service)
	if err != nil {
		return 0, err
	}

	image := fmt.Sprintf("%s/%s:%s", app, opts.Service, release.Build)

	args := []string{"run", "-i"}

	for _, v := range service.Volumes {
		args = append(args, "-v", v)
	}

	args = append(args, "--label", fmt.Sprintf("convox.app=%s", app))
	args = append(args, "--label", fmt.Sprintf("convox.release=%s", release.Id))
	args = append(args, "--label", fmt.Sprintf("convox.service=%s", opts.Service))
	args = append(args, "--link", "rack", "-e", "RACK_URL=https://rack:3000")
	args = append(args, image)

	if opts.Command != "" {
		args = append(args, "sh", "-c", opts.Command)
	}

	cmd := exec.Command("docker", args...)

	cmd.Stdin = opts.Stream
	cmd.Stdout = opts.Stream
	cmd.Stderr = opts.Stream

	err = cmd.Run()
	if ee, ok := err.(*exec.ExitError); ok {
		if status, ok := ee.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
	}

	return 0, err
}

func (p *Provider) ProcessStop(app, pid string) error {
	return exec.Command("docker", "stop", pid).Run()
}
