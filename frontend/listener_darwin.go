package frontend

import (
	"fmt"
	"os"
	"os/exec"
)

func createListener(name, subnet string) (string, error) {
	cmd := exec.Command("ifconfig", name, "create")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil
	}

	ip := fmt.Sprintf("%s.0", subnet)

	cmd = exec.Command("ifconfig", name, ip, "netmask", "255.255.255.255", "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil
	}

	return ip, nil
}

func destroyListener(name string) error {
	return exec.Command("ifconfig", name, "destroy").Run()
}
