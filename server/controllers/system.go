package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/types"
)

func SystemGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	system, err := Provider.SystemGet()
	if err != nil {
		return err
	}

	return c.RenderJSON(system)
}

func SystemLogs(w http.ResponseWriter, r *http.Request, c *api.Context) error {
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

	logs, err := Provider.SystemLogs(opts)
	if err != nil {
		return err
	}

	w.WriteHeader(200)

	if err := stream(w, logs); err != nil {
		return err
	}

	return nil
}

func SystemOptions(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	options, err := Provider.SystemOptions()
	if err != nil {
		return err
	}

	return c.RenderJSON(options)
}

func SystemUpdate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	password := c.Form("password")
	version := c.Form("version")

	opts := types.SystemUpdateOptions{
		Password: password,
		Version:  version,
	}

	if err := Provider.SystemUpdate(opts); err != nil {
		return err
	}

	w.Write([]byte("ok"))

	return nil
}
