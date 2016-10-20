package client

import (
	"fmt"

	"github.com/convox/praxis/provider/models"
)

func (c *Client) BuildCreate(app, url string, opts models.BuildCreateOptions) (*models.Build, error) {
	var build models.Build

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

func (c *Client) BuildGet(app, id string) (*models.Build, error) {
	var build models.Build

	if err := c.Get(fmt.Sprintf("/apps/%s/builds/%s", app, id), &build, GetOptions{}); err != nil {
		return nil, err
	}

	return &build, nil
}
