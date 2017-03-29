package frontend

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/convox/logger"
)

var (
	Log = logger.New("ns=frontend")
)

func Serve(iface, subnet string) error {
	log := Log.At("serve").Namespace("interface=%s subnet=%q", iface, subnet)

	ip, err := setupListener(iface, subnet)
	if err != nil {
		log.Error(err)
		return err
	}

	go startDns("convox", ip)
	go startApi(ip, iface, subnet)

	log.Success()

	select {}
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}
