package frontend

import (
	"fmt"
)

func (d *DNS) setupResolver(domain string) error {
	data := []byte(fmt.Sprintf("nameserver %s\nport 53\n", d.Host))

	if err := writeFile(fmt.Sprintf("/etc/resolver/%s", domain), data); err != nil {
		return err
	}

	return nil
}
