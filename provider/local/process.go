package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func (p *Provider) ProcessGet(app, pid string) (*types.Process, error) {
	data, err := exec.Command("docker", "inspect", pid, "--format", "{{.ID}}").CombinedOutput()
	if err != nil {
		return nil, err
	}

	fpid := strings.TrimSpace(string(data))

	filters := []string{
		fmt.Sprintf("label=convox.app=%s", app),
		fmt.Sprintf("id=%s", fpid),
	}

	fmt.Printf("filters = %+v\n", filters)

	pss, err := processList(filters)
	if err != nil {
		return nil, err
	}

	if len(pss) != 1 {
		return nil, fmt.Errorf("no such process: %s", pid)
	}

	return &pss[0], nil
}

func (p *Provider) ProcessList(app string, opts types.ProcessListOptions) (types.Processes, error) {
	filters := []string{fmt.Sprintf("label=convox.app=%s", app)}

	if opts.Service != "" {
		filters = append(filters, fmt.Sprintf("label=convox.service=%s", opts.Service))
	}

	return processList(filters)
}

func processList(filters []string) (types.Processes, error) {
	args := []string{"ps"}

	for _, f := range filters {
		args = append(args, "--filter", f)
	}

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

func (p *Provider) ProcessLogs(app, pid string) (io.ReadCloser, error) {
	_, err := p.ProcessGet(app, pid)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	fmt.Printf("w = %+v\n", w)
	cmd := exec.Command("docker", "logs", "-f", pid)

	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		cmd.Wait()
		w.Close()
	}()

	return r, nil
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

	for k, v := range opts.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
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

	if opts.Stream != nil {
		cmd.Stdin = opts.Stream
		cmd.Stdout = opts.Stream
		cmd.Stderr = opts.Stream
	}

	err = cmd.Run()

	if ee, ok := err.(*exec.ExitError); ok {
		if status, ok := ee.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
	}

	return 0, err
}

func (p *Provider) ProcessStart(app string, opts types.ProcessStartOptions) (string, error) {
	release, err := p.ReleaseGet(app, opts.Release)
	if err != nil {
		return "", err
	}

	build, err := p.BuildGet(app, release.Build)
	if err != nil {
		return "", err
	}

	m, err := manifest.Load([]byte(build.Manifest))
	if err != nil {
		return "", err
	}

	service, err := m.Services.Find(opts.Service)
	if err != nil {
		return "", err
	}

	image := fmt.Sprintf("%s/%s:%s", app, opts.Service, release.Build)

	args := []string{"run", "-i", "--detach"}

	for _, v := range service.Volumes {
		args = append(args, "-v", v)
	}

	for k, v := range opts.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, "--label", fmt.Sprintf("convox.app=%s", app))
	args = append(args, "--label", fmt.Sprintf("convox.release=%s", release.Id))
	args = append(args, "--label", fmt.Sprintf("convox.service=%s", opts.Service))
	args = append(args, "--link", "rack", "-e", "RACK_URL=https://rack:3000")
	args = append(args, image)

	if opts.Command != "" {
		args = append(args, "sh", "-c", opts.Command)
	}

	data, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func (p *Provider) ProcessStop(app, pid string) error {
	return exec.Command("docker", "stop", "-t", "2", pid).Run()
}
