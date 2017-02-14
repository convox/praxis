package controllers

import (
	"net/http"

	"github.com/convox/praxis/api"
)

func AppCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	name := c.Form("name")

	c.LogParams("name")

	app, err := Provider.AppCreate(name)
	if err != nil {
		return err
	}

	return c.RenderJSON(app)
}
