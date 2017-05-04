package rack

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/convox/praxis/types"
)

func (c *Client) ReleaseCreate(app string, opts types.ReleaseCreateOptions) (release *types.Release, err error) {
	data, err := json.Marshal(opts.Env)
	if err != nil {
		return nil, err
	}

	ro := RequestOptions{
		Params: Params{
			"build": opts.Build,
			"env":   string(data),
		},
	}

	err = c.Post(fmt.Sprintf("/apps/%s/releases", app), ro, &release)

	return
}

func (c *Client) ReleaseGet(app, id string) (release *types.Release, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/releases/%s", app, id), RequestOptions{}, &release)
	return
}

func (c *Client) ReleaseList(app string, opts types.ReleaseListOptions) (releases types.Releases, err error) {
	ro := RequestOptions{Query: Query{}}

	if opts.Count > 0 {
		ro.Query["count"] = strconv.Itoa(opts.Count)
	}

	err = c.Get(fmt.Sprintf("/apps/%s/releases", app), ro, &releases)
	return
}

func (c *Client) ReleaseLogs(app, id string) (io.ReadCloser, error) {
	res, err := c.GetStream(fmt.Sprintf("/apps/%s/releases/%s/logs", app, id), RequestOptions{})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) ReleasePromote(app, id string) error {
	return c.Post(fmt.Sprintf("/apps/%s/releases/%s/promote", app, id), RequestOptions{}, nil)
}
