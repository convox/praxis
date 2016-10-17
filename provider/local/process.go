package local

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/convox/praxis/provider/models"
)

func (p *Provider) ProcessRun(app, service string, opts models.ProcessRunOptions) (*models.Process, error) {
	cmd := exec.Command("docker", append([]string{"run", "-i", service}, opts.Command...)...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	ps := &models.Process{
		Id: fmt.Sprintf("%d", cmd.Process.Pid),
	}

	return ps, nil
}
