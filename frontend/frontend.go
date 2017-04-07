package frontend

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/convox/logger"
)

var (
	Host      string
	Interface string
	Log       = logger.New("ns=frontend")
)

type Frontend struct {
	API       *API
	DNS       *DNS
	Interface string
	Subnet    string

	domains   map[string]bool
	endpoints map[string]Endpoint
	hosts     map[string]string
	logger    *logger.Logger
}

func New(iface, subnet string) *Frontend {
	return &Frontend{
		Interface: iface,
		Subnet:    subnet,
		domains:   map[string]bool{},
		endpoints: map[string]Endpoint{},
		hosts:     map[string]string{},
		logger:    logger.New("ns=frontend"),
	}
}

func (f *Frontend) Serve() error {
	log := f.logger.At("serve").Namespace("interface=%s subnet=%q", f.Interface, f.Subnet)

	ip, err := setupListener(f.Interface, f.Subnet)
	if err != nil {
		log.Error(err)
		return err
	}

	f.API = NewAPI(ip, f)
	f.DNS = NewDNS(ip, f)

	go f.API.Serve()
	go f.DNS.Serve()

	log.Success()
	select {}
}

func (f *Frontend) nextHostIP() (string, error) {
	for i := 1; i < 255; i++ {
		ip := fmt.Sprintf("%s.%d", f.Subnet, i)

		found := false

		for _, hip := range f.hosts {
			if ip == hip {
				found = true
				break
			}
		}

		if !found {
			return ip, nil
		}
	}

	return "", fmt.Errorf("ip space exhausted")
}

func execute(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}
