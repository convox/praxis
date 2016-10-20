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
