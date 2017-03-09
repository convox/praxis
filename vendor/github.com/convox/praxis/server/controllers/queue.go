package controllers

import (
	"net/http"

	"github.com/convox/api"
	"github.com/convox/praxis/types"
)

func QueueFetch(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	queue := c.Var("queue")

	attrs, err := Provider.QueueFetch(app, queue, types.QueueFetchOptions{})
	if err != nil {
		return err
	}

	return c.RenderJSON(attrs)
}

func QueueStore(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	queue := c.Var("queue")

	if err := r.ParseForm(); err != nil {
		return err
	}

	attrs := map[string]string{}

	for k := range r.Form {
		attrs[k] = r.Form.Get(k)
	}

	err := Provider.QueueStore(app, queue, attrs)
	if err != nil {
		return err
	}

	return c.RenderJSON("")
}
