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

func TableFetch(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")
	index := c.Var("index")
	key := c.Var("key")

	attrs, err := Provider.TableFetch(app, table, key, types.TableFetchOptions{Index: index})
	if err != nil {
		return err
	}

	return c.RenderJSON(attrs)
}

func TableFetchBatch(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	app := c.Var("app")
	table := c.Var("table")
	index := c.Var("index")
	keys := r.Form["key"]

	items, err := Provider.TableFetchBatch(app, table, keys, types.TableFetchOptions{Index: index})
	if err != nil {
		return err
	}

	return c.RenderJSON(items)
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

func TableRemove(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")
	index := c.Var("index")
	key := c.Var("key")

	return Provider.TableRemove(app, table, key, types.TableRemoveOptions{Index: index})
}

func TableRemoveBatch(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	app := c.Var("app")
	table := c.Var("table")
	index := c.Var("index")
	keys := r.Form["key"]

	return Provider.TableRemoveBatch(app, table, keys, types.TableRemoveOptions{Index: index})
}

func TableTruncate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")

	return Provider.TableTruncate(app, table)
}
