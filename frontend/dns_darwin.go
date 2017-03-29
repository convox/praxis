package frontend

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func setupResolver(root, ip string) error {
	path := filepath.Join("/etc", "resolver", root)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(fmt.Sprintf("nameserver %s\nport 53\n", ip)), 0644)
}
