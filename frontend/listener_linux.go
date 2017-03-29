package frontend

import (
	"fmt"
	"os"
	"os/exec"
)

func createListener(name, subnet string) (string, error) {
	cmd := exec.Command("ip", "link", "add", "link", "docker0", "name", name, "type", "vlan", "id", "1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil
	}

	ip := fmt.Sprintf("%s.0", subnet)

	cmd = exec.Command("ip", "addr", "add", ip, "dev", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil
	}

	return ip, nil
}

func destroyListener(name string) error {
	return exec.Command("ip", "link", "del", name).Run()
}
