package local

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/convox/praxis/provider"
)

func (p *Provider) ProcessStart(app, service string, opts provider.ProcessRunOptions) (*provider.Process, error) {
	args := []string{"run", "-i", "-d"}

	for k, v := range opts.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// special for builder
	args = append(args, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	args = append(args, "--network", "host")
	args = append(args, "-e", "CONVOX_URL=https://localhost:9877")

	args = append(args, "--label", fmt.Sprintf("convox.system=%s", SystemId))
	args = append(args, "--label", fmt.Sprintf("convox.app=%s", app))

	args = append(args, service)
	args = append(args, opts.Command...)

	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return nil, fmt.Errorf(strings.Split(string(out), "\n")[0])
		}

		return nil, err
	}

	ps := &provider.Process{
		Id: strings.TrimSpace(string(out)),
	}

	return ps, nil
}

func (p *Provider) ProcessWait(app, pid string) (int, error) {
	out, err := exec.Command("docker", "wait", pid).CombinedOutput()
	if err != nil {
		return -1, err
	}

	status, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return -1, err
	}

	return status, nil
}
