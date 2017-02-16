package controllers

import (
	"net/http"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/types"
)

func BuildCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	c.LogParams("url", "cache")

	app := c.Var("app")
	url := c.Form("url")
	cache := c.Form("cache") == "true"

	build, err := Provider.BuildCreate(app, url, types.BuildCreateOptions{Cache: cache})
	if err != nil {
		return err
	}

	return c.RenderJSON(build)
}

func BuildGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	id := c.Var("id")

	build, err := Provider.BuildGet(app, id)
	if err != nil {
		return err
	}

	return c.RenderJSON(build)
}

func BuildLogs(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	id := c.Var("id")

	logs, err := Provider.BuildLogs(app, id)
	if err != nil {
		return err
	}

	w.WriteHeader(200)

	if err := stream(w, logs); err != nil {
		return err
	}

	return nil
}

func BuildUpdate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	c.LogParams("release", "status")

	app := c.Var("app")
	id := c.Var("id")

	manifest := c.Form("manifest")
	release := c.Form("release")
	status := c.Form("status")

	build, err := Provider.BuildUpdate(app, id, types.BuildUpdateOptions{
		Manifest: manifest,
		Release:  release,
		Status:   status,
	})
	if err != nil {
		return err
	}

	return c.RenderJSON(build)
}
