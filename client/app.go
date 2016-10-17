package client

import "github.com/convox/praxis/provider/models"

func (c *Client) AppCreate(name string, opts models.AppCreateOptions) (*models.App, error) {
	var app models.App

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
