package rack

import (
	"fmt"
	"io"

	"github.com/convox/praxis/types"
)

func (c *Client) BuildCreate(app, url string) (build *types.Build, err error) {
	ro := RequestOptions{
		Params: Params{
			"url": url,
		},
	}

	err = c.Post(fmt.Sprintf("/apps/%s/builds", app), ro, &build)
	return
}

func (c *Client) BuildGet(app, id string) (build *types.Build, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/builds/%s", app, id), RequestOptions{}, &build)
	return
}

func (c *Client) BuildLogs(app, id string) (io.ReadCloser, error) {
	res, err := c.GetStream(fmt.Sprintf("/apps/%s/builds/%s/logs", app, id), RequestOptions{})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) BuildUpdate(app, id string, opts types.BuildUpdateOptions) (build *types.Build, err error) {
	ro := RequestOptions{
		Params: Params{
			"manifest": opts.Manifest,
			"release":  opts.Release,
			"status":   opts.Status,
		},
	}

	err = c.Put(fmt.Sprintf("/apps/%s/builds/%s", app, id), ro, &build)
	return
}
