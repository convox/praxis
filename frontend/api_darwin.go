package frontend

import (
	"fmt"
	"os"
	"os/exec"
)

func createHost(iface, subnet, host string) (string, error) {
	ip := fmt.Sprintf("%s.%d", subnet, len(endpoints)+1)

	cmd := exec.Command("ifconfig", iface, "alias", ip, "netmask", "255.255.255.255")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil
	}

	endpoints[ip] = map[int]Endpoint{}
	hosts[host] = ip

	return ip, nil
}
