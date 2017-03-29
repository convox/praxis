package frontend

import (
	"fmt"
)

func createListener(name, subnet string) (string, error) {
	if err := execute("ifconfig", name, "create"); err != nil {
		return "", err
	}

	ip := fmt.Sprintf("%s.0", subnet)

	if err := execute("ifconfig", name, ip, "netmask", "255.255.255.255", "up"); err != nil {
		return "", err
	}

	return ip, nil
}

func destroyListener(name string) error {
	return execute("ifconfig", name, "destroy")
}
