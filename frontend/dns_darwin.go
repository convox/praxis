package frontend

import (
	"fmt"
)

func setupResolver(root, ip string) error {
	data := []byte(fmt.Sprintf("nameserver %s\nport 53\n", ip))

	if err := writeFile(fmt.Sprintf("/etc/resolver/%s", root), data); err != nil {
		return err
	}

	return nil
}
