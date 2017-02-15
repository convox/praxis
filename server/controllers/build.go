package controllers

import (
	"net/http"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/types"
)

func BuildCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	c.LogParams("url")

	app := c.Var("app")
	url := c.Form("url")
	cache := c.Form("cache") == "true"

	build, err := Provider.BuildCreate(app, url, types.BuildCreateOptions{Cache: cache})
	if err != nil {
		return err
	}

	return c.RenderJSON(build)
}
