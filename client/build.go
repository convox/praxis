package client

import (
	"fmt"
	"io"

	"github.com/convox/praxis/provider"
)

type Build provider.Build
type Builds provider.Builds

type BuildCreateOptions provider.BuildCreateOptions

func (c *Client) BuildCreate(app, url string, opts BuildCreateOptions) (*Build, error) {
	var build Build

	popts := PostOptions{
		Params: map[string]string{
			"url": url,
		},
	}

	if err := c.Post(fmt.Sprintf("/apps/%s/builds", app), &build, popts); err != nil {
		return nil, err
	}

	return &build, nil
}

func (c *Client) BuildGet(app, id string) (*Build, error) {
	var build Build

	if err := c.Get(fmt.Sprintf("/apps/%s/builds/%s", app, id), &build, GetOptions{}); err != nil {
		return nil, err
	}

	return &build, nil
}

func (c *Client) BuildLogs(app, id string) (io.ReadCloser, error) {
	return c.GetReader(fmt.Sprintf("/apps/%s/builds/%s/logs", app, id), GetOptions{})
}
