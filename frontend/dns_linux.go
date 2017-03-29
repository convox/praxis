package frontend

import (
	"fmt"
	"os/exec"
)

func setupResolver(root, ip string) error {
	data := []byte("[main]\ndns=dnsmasq\n")

	if err := writeFile("/etc/NetworkManager/conf.d/convox.conf", data); err != nil {
		return err
	}

	data = []byte(fmt.Sprintf("server=/%s/%s\n", root, ip))

	if err := writeFile(fmt.Sprintf("/etc/NetworkManager/dnsmasq.d/%s", root), data); err != nil {
		return err
	}

	if err := exec.Command("systemctl", "restart", "NetworkManager").Run(); err != nil {
		return err
	}

	return nil
}
