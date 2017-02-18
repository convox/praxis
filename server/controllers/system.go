package controllers

import (
	"net/http"

	"github.com/convox/praxis/api"
)

func SystemGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	system, err := Provider.SystemGet()
	if err != nil {
		return err
	}

	return c.RenderJSON(system)
}
