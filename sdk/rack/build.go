package rack

import (
	"fmt"
	"io"

	"github.com/convox/praxis/types"
)

func (c *Client) BuildCreate(app, url string) (build *types.Build, err error) {
	err = c.Post(fmt.Sprintf("/apps/%s/builds", app), Params{"url": url}, &build)
	return
}

func (c *Client) BuildGet(app, id string) (build *types.Build, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/builds/%s", app, id), &build)
	return
}

func (c *Client) BuildLogs(app, id string) (io.Reader, error) {
	return c.GetStream(fmt.Sprintf("/apps/%s/builds/%s/logs", app, id))
}

func (c *Client) BuildUpdate(app, id string, opts types.BuildUpdateOptions) (build *types.Build, err error) {
	params := Params{
		"manifest": opts.Manifest,
		"release":  opts.Release,
		"status":   opts.Status,
	}

	err = c.Put(fmt.Sprintf("/apps/%s/builds/%s", app, id), params, &build)
	return
}
