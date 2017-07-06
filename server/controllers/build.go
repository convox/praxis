package controllers

import (
	"net/http"
	"sort"
	"time"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
)

func BuildCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	cache := c.Form("cache") == "true"
	development := c.Form("development") == "true"
	url := c.Form("url")

	opts := types.BuildCreateOptions{
		Cache:       cache,
		Development: development,
	}

	build, err := Provider.WithContext(c.Context()).BuildCreate(app, url, opts)
	if err != nil {
		return err
	}

	return c.RenderJSON(build)
}

func BuildGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	id := c.Var("id")

	if _, err := Provider.WithContext(c.Context()).AppGet(app); err != nil {
		return err
	}

	build, err := Provider.BuildGet(app, id)
	if err != nil {
		return err
	}

	return c.RenderJSON(build)
}

func BuildList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

	builds, err := Provider.BuildList(app)
	if err != nil {
		return err
	}

	sort.Slice(builds, func(i, j int) bool { return builds[j].Created.Before(builds[i].Created) })

	return c.RenderJSON(builds)
}

func BuildLogs(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	id := c.Var("id")

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

	logs, err := Provider.BuildLogs(app, id)
	if err != nil {
		return err
	}

	w.WriteHeader(200)

	err = helpers.Stream(w, logs)
	return err
}

func BuildUpdate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	id := c.Var("id")

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

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
