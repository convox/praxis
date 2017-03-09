package local

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
)

func (p *Provider) Proxy(app, pid string, port int, in io.Reader) (io.ReadCloser, error) {
	_, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	data, err := exec.Command("docker", "inspect", pid, "--format", "{{.NetworkSettings.IPAddress}}").CombinedOutput()
	if err != nil {
		return nil, err
	}

	ip := strings.TrimSpace(string(data))

	cn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err
	}

	go io.Copy(cn, in)

	return cn, nil
}
