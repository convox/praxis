package frontend

import (
	"fmt"
)

func (f *Frontend) createHost(host string) (string, error) {
	if _, ok := f.hosts[host]; ok {
		return "", fmt.Errorf("host %s already exists", host)
	}

	ip, err := f.nextHostIP()
	if err != nil {
		return "", err
	}

	if err := execute("ip", "addr", "add", ip, "dev", f.Interface); err != nil {
		return "", err
	}

	f.hosts[host] = ip

	return ip, nil
}
