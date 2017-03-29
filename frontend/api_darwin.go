package frontend

import (
	"fmt"
)

func createHost(iface, subnet, host string) (string, error) {
	ip := fmt.Sprintf("%s.%d", subnet, len(endpoints)+1)

	if err := execute("ifconfig", iface, "alias", ip, "netmask", "255.255.255.255"); err != nil {
		return "", err
	}

	endpoints[ip] = map[int]Endpoint{}
	hosts[host] = ip

	return ip, nil
}
