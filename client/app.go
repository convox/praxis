package client

import (
	"fmt"

	"github.com/convox/praxis/provider"
)

type App provider.App
type Apps provider.Apps

type AppCreateOptions provider.AppCreateOptions

func (c *Client) AppCreate(name string, opts AppCreateOptions) (*App, error) {
	var app App

	popts := PostOptions{
		Params: map[string]string{
			"name": name,
		},
	}

	if err := c.Post("/apps", &app, popts); err != nil {
		return nil, err
	}

	return &app, nil
}

func (c *Client) AppDelete(name string) error {
	return c.Delete(fmt.Sprintf("/apps/%s", name), DeleteOptions{})
}
