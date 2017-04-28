package controllers

import (
	"net/http"

	"github.com/convox/api"
	"github.com/convox/praxis/types"
)

func SystemGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	system, err := Provider.SystemGet()
	if err != nil {
		return err
	}

	return c.RenderJSON(system)
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
