package controllers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/types"
)

func BuildCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	c.LogParams("url", "cache")

	app := c.Var("app")
	url := c.Form("url")
	cache := c.Form("cache") == "true"

	opts := types.BuildCreateOptions{
		Cache: cache,
	}

	if s, err := strconv.Atoi(c.Form("stage")); err == nil {
		opts.Stage = s
	}

	build, err := Provider.BuildCreate(app, url, opts)
	if err != nil {
		return err
	}

	return c.RenderJSON(build)
}

func BuildGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	id := c.Var("id")

	if _, err := Provider.AppGet(app); err != nil {
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

	if err := stream(w, logs); err != nil {
		return err
	}

	return nil
}

func BuildUpdate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	c.LogParams("release", "status")

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
