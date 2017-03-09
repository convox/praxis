package local

import (
	"fmt"
	"io"
	"os/exec"
)

func (p *Provider) FilesDelete(app, pid string, files []string) error {
	args := []string{"exec", pid, "rm", "-f"}
	args = append(args, files...)

	return exec.Command("docker", args...).Run()
}

func (p *Provider) FilesUpload(app, pid string, r io.Reader) error {
	cmd := exec.Command("docker", "cp", "-", fmt.Sprintf("%s:.", pid))

	cmd.Stdin = r

	return cmd.Run()
}
