package rack

import (
	"fmt"
	"io"
	"strconv"

	"github.com/convox/praxis/types"
)

func (c *Client) AppCreate(name string) (app *types.App, err error) {
	ro := RequestOptions{
		Params: Params{
			"name": name,
		},
	}

	err = c.Post("/apps", ro, &app)
	return
}

func (c *Client) AppDelete(name string) error {
	return c.Delete(fmt.Sprintf("/apps/%s", name), RequestOptions{}, nil)
}

func (c *Client) AppGet(name string) (app *types.App, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s", name), RequestOptions{}, &app)
	return
}

func (c *Client) AppList() (apps types.Apps, err error) {
	err = c.Get("/apps", RequestOptions{}, &apps)
	return
}

func (c *Client) AppLogs(app string, opts types.AppLogsOptions) (io.ReadCloser, error) {
	ro := RequestOptions{
		Query: Query{
			"filter": opts.Filter,
			"follow": fmt.Sprintf("%t", opts.Follow),
		},
	}

	if !opts.Since.IsZero() {
		ro.Query["since"] = strconv.Itoa(int(opts.Since.UTC().Unix()))
	}

	res, err := c.GetStream(fmt.Sprintf("/apps/%s/logs", app), ro)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) AppRegistry(app string) (registry *types.Registry, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/registry", app), RequestOptions{}, &registry)
	return
}
