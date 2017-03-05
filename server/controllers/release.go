package controllers

import (
	"net/http"

	"github.com/convox/api"
	"github.com/convox/praxis/types"
)

func ReleaseCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	build := c.Form("build")
	env := map[string]string{}

	c.LogParams("build")

	release, err := Provider.ReleaseCreate(app, types.ReleaseCreateOptions{
		Build: build,
		Env:   env,
	})
	if err != nil {
		return err
	}

	return c.RenderJSON(release)
}

func ReleaseGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	id := c.Var("id")

	release, err := Provider.ReleaseGet(app, id)
	if err != nil {
		return err
	}

	return c.RenderJSON(release)
}
