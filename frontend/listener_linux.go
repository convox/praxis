package frontend

import (
	"fmt"
)

func createListener(name, subnet string) (string, error) {
	if err := execute("ip", "link", "add", "link", "docker0", "name", name, "type", "vlan", "id", "1"); err != nil {
		return "", err
	}

	ip := fmt.Sprintf("%s.0", subnet)

	if err := execute("ip", "addr", "add", ip, "dev", name); err != nil {
		return "", err
	}

	return ip, nil
}

func destroyListener(name string) error {
	return execute("ip", "link", "del", name)
}
