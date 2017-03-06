package controllers

import (
	"net/http"

	"github.com/convox/api"
)

func TableFetch(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")
	id := c.Var("id")

	attrs, err := Provider.TableFetch(app, table, id)
	if err != nil {
		return err
	}

	return c.RenderJSON(attrs)
}

func TableStore(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")

	if err := r.ParseForm(); err != nil {
		return err
	}

	attrs := map[string]string{}

	for k := range r.Form {
		attrs[k] = r.Form.Get(k)
	}

	id, err := Provider.TableStore(app, table, attrs)
	if err != nil {
		return err
	}

	return c.RenderJSON(id)
}
