package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) SystemGet() (system *types.System, err error) {
	err = c.Get("/system", RequestOptions{}, &system)
	return
}

func (c *Client) SystemInstall(name string, opts types.SystemInstallOptions) (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (c *Client) SystemUninstall(name string, opts types.SystemInstallOptions) error {
	return fmt.Errorf("unimplemented")
}

func (c *Client) SystemUpdate(opts types.SystemUpdateOptions) error {
	ro := RequestOptions{
		Params: Params{
			"version": opts.Version,
		},
	}

	return c.Post("/system", ro, nil)
}
