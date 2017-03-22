package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) RegistryAdd(hostname, username, password string) (registry *types.Registry, err error) {
	ro := RequestOptions{
		Params: Params{
			"hostname": hostname,
			"username": username,
			"password": password,
		},
	}

	err = c.Post("/registries", ro, &registry)
	return
}

func (c *Client) RegistryList() (registries types.Registries, err error) {
	err = c.Get("/registries", RequestOptions{}, &registries)
	return
}

func (c *Client) RegistryRemove(hostname string) error {
	return c.Delete(fmt.Sprintf("/registries/%s", hostname), RequestOptions{}, nil)
}
