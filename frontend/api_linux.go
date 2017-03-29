package frontend

import (
	"fmt"
)

func createHost(iface, subnet, host string) (string, error) {
	ip := fmt.Sprintf("%s.%d", subnet, len(endpoints)+1)

	if err := execute("ip", "addr", "add", ip, "dev", iface); err != nil {
		return "", err
	}

	endpoints[ip] = map[int]Endpoint{}
	hosts[host] = ip

	return ip, nil
}
