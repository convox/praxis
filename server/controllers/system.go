package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/helpers"
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

	if err := helpers.Stream(w, logs); err != nil {
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

// func SystemProxy(rw io.ReadWriteCloser, c *api.Context) error {
//   host := c.Var("host")
//   port := c.Var("port")

//   pi, err := strconv.Atoi(port)
//   if err != nil {
//     return err
//   }

//   p, err := Provider.SystemProxy(host, pi, rw)
//   if err != nil {
//     return err
//   }

//   defer p.Close()

//   if _, err := io.Copy(rw, p); err != nil {
//     return err
//   }

//   return nil
// }

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
