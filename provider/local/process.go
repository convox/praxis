package local

import (
	"fmt"
	"os/exec"

	"github.com/convox/praxis/types"
)

func (p *Provider) ProcessRun(app string, opts types.ProcessRunOptions) error {
	image := fmt.Sprintf("%s/%s", app, opts.Service)

	args := []string{"run"}
	args = append(args, "-i")
	args = append(args, "--link", "rack", "-e", "RACK_URL=https://rack:3000")
	args = append(args, image)

	if opts.Command != "" {
		args = append(args, "sh", "-c", opts.Command)
	}

	cmd := exec.Command("docker", args...)

	cmd.Stdin = opts.Stream
	cmd.Stdout = opts.Stream
	cmd.Stderr = opts.Stream

	return cmd.Run()
}
