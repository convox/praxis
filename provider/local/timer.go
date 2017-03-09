package local

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/convox/praxis/types"
)

func (p *Provider) TimerCreate(app, name string, opts types.TimerCreateOptions) error {
	cname := fmt.Sprintf("timer-%s-%s", app, name)

	exec.Command("docker", "rm", "-f", cname).Run()

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	args := []string{"run", "-i"}

	// TODO: use real app env
	args = append(args, "-e", fmt.Sprintf("APP=%s", app))

	args = append(args, "--link", hostname, "-e", fmt.Sprintf("RACK_URL=https://%s:3000", hostname))
	args = append(args, "--name", cname)
	args = append(args, "convox/praxis", "timer")
	args = append(args, "-app", app)
	args = append(args, "-name", name)
	args = append(args, "-command", opts.Command)
	args = append(args, "-schedule", opts.Schedule)
	args = append(args, "-service", opts.Service)

	cmd := exec.Command("docker", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}
