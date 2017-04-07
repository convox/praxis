package frontend

import (
	"fmt"
)

func (d *DNS) setupResolver(domain string) error {
	data := []byte("[main]\ndns=dnsmasq\n")

	if err := writeFile("/etc/NetworkManager/conf.d/convox.conf", data); err != nil {
		return err
	}

	data = []byte(fmt.Sprintf("server=/%s/%s\n", domain, d.Host))

	if err := writeFile(fmt.Sprintf("/etc/NetworkManager/dnsmasq.d/%s", domain), data); err != nil {
		return err
	}

	if err := execute("systemctl", "restart", "NetworkManager"); err != nil {
		return err
	}

	return nil
}
