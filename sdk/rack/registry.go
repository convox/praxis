package rack

import (
	"github.com/convox/praxis/types"
)

func (c *Client) RegistryAdd(server, username, password string) (registry *types.Registry, err error) {
	ro := RequestOptions{
		Params: Params{
			"server":   server,
			"username": username,
			"password": password,
		},
	}

	err = c.Post("/registries", ro, &registry)
	return
}

func (c *Client) RegistryDelete(server string) error {
	ro := RequestOptions{
		Params: Params{
			"server": server,
		},
	}

	return c.Delete("/registries", ro, nil)
}

func (c *Client) RegistryList() (registries types.Registries, err error) {
	err = c.Get("/registries", RequestOptions{}, &registries)
	return
}
