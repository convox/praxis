package local

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/convox/praxis/provider/models"
)

func (p *Provider) ProcessStart(app, service string, opts models.ProcessRunOptions) (*models.Process, error) {
	args := []string{"run", "-i"}

	args = append(args, "-v", "/var/run/convox:/var/run/convox")

	for k, v := range opts.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, service)
	args = append(args, opts.Command...)

	fmt.Printf("args = %+v\n", args)

	cmd := exec.Command("docker", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if len(opts.Environment) > 0 {
		cmd.Env = []string{}

		for k, v := range opts.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	ps := &models.Process{
		Id: fmt.Sprintf("%d", cmd.Process.Pid),
	}

	return ps, nil
}

func (p *Provider) ProcessWait(app, pid string) (int, error) {
	ipid, err := strconv.Atoi(pid)
	if err != nil {
		return -1, err
	}

	ps, err := os.FindProcess(ipid)
	if err != nil {
		return -1, err
	}

	status, err := ps.Wait()
	if err != nil {
		return -1, err
	}

	sw, ok := status.Sys().(syscall.WaitStatus)
	fmt.Printf("sw = %+v\n", sw)
	fmt.Printf("ok = %+v\n", ok)
	if !ok {
		return -1, fmt.Errorf("could not get exit status")
	}

	fmt.Printf("sw.ExitStatus() = %+v\n", sw.ExitStatus())

	return sw.ExitStatus(), nil
}
