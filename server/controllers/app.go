package controllers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/convox/api"
	"github.com/convox/praxis/types"
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

	sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })

	return c.RenderJSON(apps)
}

func AppLogs(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

	opts := types.LogsOptions{
		Filter: c.Query("filter"),
		Follow: c.Query("follow") == "true",
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

	if err := stream(w, logs); err != nil {
		return err
	}

	return nil
}

func AppRegistry(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	registry, err := Provider.AppRegistry(app)
	if err != nil {
		return err
	}

	return c.RenderJSON(registry)
}
