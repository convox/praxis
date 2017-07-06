package controllers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
)

func AppCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app, err := Provider.WithContext(c.Context()).AppCreate(c.Form("name"))
	if err != nil {
		return err
	}

	return c.RenderJSON(app)
}

func AppDelete(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	if err := Provider.WithContext(c.Context()).AppDelete(c.Var("name")); err != nil {
		return err
	}

	return c.RenderOK()
}

func AppGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app, err := Provider.WithContext(c.Context()).AppGet(c.Var("name"))
	if err != nil {
		return err
	}

	return c.RenderJSON(app)
}

func AppList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	apps, err := Provider.WithContext(c.Context()).AppList()
	if err != nil {
		return err
	}

	sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })

	return c.RenderJSON(apps)
}

func AppLogs(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	if _, err := Provider.WithContext(c.Context()).AppGet(app); err != nil {
		return err
	}

	opts := types.LogsOptions{
		Filter: c.Query("filter"),
		Follow: c.Query("follow") == "true",
		Prefix: c.Query("prefix") == "true",
	}

	if since := c.Query("since"); since != "" {
		t, err := strconv.Atoi(since)
		if err != nil {
			return err
		}

		opts.Since = time.Unix(int64(t), 0)
	}

	logs, err := Provider.AppLogs(app, opts)
	if err != nil {
		return err
	}

	w.WriteHeader(200)

	err = helpers.Stream(w, logs)
	return err
}

func AppRegistry(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	registry, err := Provider.WithContext(c.Context()).AppRegistry(c.Var("app"))
	if err != nil {
		return err
	}

	return c.RenderJSON(registry)
}
