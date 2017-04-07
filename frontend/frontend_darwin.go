package frontend

import "fmt"

func (f *Frontend) createHost(host string) (string, error) {
	if _, ok := f.hosts[host]; ok {
		return "", fmt.Errorf("host %s already exists", host)
	}

	ip, err := f.nextHostIP()
	if err != nil {
		return "", err
	}

	if err := execute("ifconfig", f.Interface, "alias", ip, "netmask", "255.255.255.255"); err != nil {
		return "", err
	}

	f.hosts[host] = ip

	return ip, nil
}
