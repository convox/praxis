package frontend

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}
