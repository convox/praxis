package controllers

import (
	"net/http"

	"github.com/convox/api"
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

func AppDelete(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	name := c.Var("name")

	if err := Provider.AppDelete(name); err != nil {
		return err
	}

	return nil
}

func AppGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	name := c.Var("name")

	app, err := Provider.AppGet(name)
	if err != nil {
		return err
	}

	return c.RenderJSON(app)
}

func AppList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	apps, err := Provider.AppList()
	if err != nil {
		return err
	}

	return c.RenderJSON(apps)
}

func AppLogs(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

	logs, err := Provider.AppLogs(app)
	if err != nil {
		return err
	}

	w.WriteHeader(200)

	if err := stream(w, logs); err != nil {
		return err
	}

	return nil
}
