package local

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
)

func (p *Provider) ProxyStart(app, pid string, port int) (io.ReadWriter, error) {
	_, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	data, err := exec.Command("docker", "inspect", pid, "--format", "{{.NetworkSettings.IPAddress}}").CombinedOutput()
	if err != nil {
		return nil, err
	}

	ip := strings.TrimSpace(string(data))

	return net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
}
