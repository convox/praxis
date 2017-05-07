package controllers

import (
	"net/http"
	"sort"

	"github.com/convox/praxis/api"
)

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

	sort.Slice(t, t.Less)

	return c.RenderJSON(t)
}

func TableTruncate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	table := c.Var("table")

	return Provider.TableTruncate(app, table)
}
