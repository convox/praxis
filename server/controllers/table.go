package controllers

import (
	"net/http"

	"github.com/convox/api"
)

func TableFetch(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")
	index := c.Var("index")

	if err := r.ParseForm(); err != nil {
		return err
	}

	if len(r.Form["id"]) == 0 {
		return api.Errorf(400, "no id provided")
	}

	if index == "id" {
		attrs, err := Provider.TableFetch(app, table, r.Form.Get("id"))
		if err != nil {
			return err
		}

		return c.RenderJSON([]map[string]string{attrs})
	}

	var attrs []map[string]string
	var err error
	if len(r.Form["id"]) == 1 {
		attrs, err = Provider.TableFetchIndex(app, table, index, r.Form.Get("id"))
		if err != nil {
			return err
		}

	} else {
		attrs, err = Provider.TableFetchIndexBatch(app, table, index, r.Form["id"])
		if err != nil {
			return err
		}
	}

	return c.RenderJSON(attrs)
}

func TableGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")

	t, err := Provider.TableGet(app, table)
	if err != nil {
		return err
	}

	return c.RenderJSON(t)
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
