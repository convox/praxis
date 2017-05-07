package rack

import (
	"fmt"
	"io"
	"strconv"

	"github.com/convox/praxis/types"
)

func (c *Client) SystemGet() (system *types.System, err error) {
	err = c.Get("/system", RequestOptions{}, &system)
	return
}

func (c *Client) SystemInstall(name string, opts types.SystemInstallOptions) (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (c *Client) SystemLogs(opts types.LogsOptions) (io.ReadCloser, error) {
	ro := RequestOptions{
		Query: Query{
			"filter": opts.Filter,
			"follow": fmt.Sprintf("%t", opts.Follow),
			"prefix": fmt.Sprintf("%t", opts.Prefix),
		},
	}

	if !opts.Since.IsZero() {
		ro.Query["since"] = strconv.Itoa(int(opts.Since.UTC().Unix()))
	}

	res, err := c.GetStream("/system/logs", ro)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) SystemOptions() (options map[string]string, err error) {
	err = c.Options("/system", RequestOptions{}, &options)
	return
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
