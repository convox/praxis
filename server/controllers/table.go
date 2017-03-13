package controllers

import (
	"net/http"

	"github.com/convox/api"
	"github.com/convox/praxis/types"
)

func TableCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	app := c.Var("app")
	table := c.Var("table")
	indexes := r.Form["index"]

	return Provider.TableCreate(app, table, types.TableCreateOptions{Indexes: indexes})
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

func TableList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	t, err := Provider.TableList(app)
	if err != nil {
		return err
	}

	return c.RenderJSON(t)
}

func TableRowDelete(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")
	index := c.Var("index")
	key := c.Var("key")

	return Provider.TableRowDelete(app, table, key, types.TableRowDeleteOptions{Index: index})
}

func TableRowGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")
	index := c.Var("index")
	key := c.Var("key")

	attrs, err := Provider.TableRowGet(app, table, key, types.TableRowGetOptions{Index: index})
	if err != nil {
		return err
	}

	return c.RenderJSON(attrs)
}

func TableRowStore(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")

	if err := r.ParseForm(); err != nil {
		return err
	}

	attrs := map[string]string{}

	for k := range r.Form {
		attrs[k] = r.Form.Get(k)
	}

	id, err := Provider.TableRowStore(app, table, attrs)
	if err != nil {
		return err
	}

	return c.RenderJSON(id)
}

func TableRowsDelete(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	app := c.Var("app")
	table := c.Var("table")
	index := c.Var("index")
	keys := r.Form["key"]

	return Provider.TableRowsDelete(app, table, keys, types.TableRowDeleteOptions{Index: index})
}

func TableRowsGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	app := c.Var("app")
	table := c.Var("table")
	index := c.Var("index")
	keys := r.Form["key"]

	items, err := Provider.TableRowsGet(app, table, keys, types.TableRowGetOptions{Index: index})
	if err != nil {
		return err
	}

	return c.RenderJSON(items)
}

func TableTruncate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")

	return Provider.TableTruncate(app, table)
}
