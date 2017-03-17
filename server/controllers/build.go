package controllers

import (
	"net/http"
	"time"

	"github.com/convox/api"
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

func BuildList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	builds, err := Provider.BuildList(app)
	if err != nil {
		return err
	}

	return c.RenderJSON(builds)
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

	var started, ended time.Time
	var err error

	if date := c.Form("started"); date != "" {
		started, err = time.Parse(sortableTime, date)
		if err != nil {
			return err
		}
	}

	if date := c.Form("ended"); date != "" {
		ended, err = time.Parse(sortableTime, date)
		if err != nil {
			return err
		}
	}

	build, err := Provider.BuildUpdate(app, id, types.BuildUpdateOptions{
		Ended:    ended,
		Manifest: manifest,
		Release:  release,
		Started:  started,
		Status:   status,
	})
	if err != nil {
		return err
	}

	return c.RenderJSON(build)
}
