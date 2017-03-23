package controllers

import (
	"net/http"

	"github.com/convox/api"
	"github.com/convox/praxis/types"
)

func MetricGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	ns := c.Var("namespace")
	m := c.Var("metric")

	metrics, err := Provider.MetricGet(app, ns, m, types.MetricGetOptions{})
	if err != nil {
		return err
	}

	return c.RenderJSON(metrics)
}

func MetricList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	ns := c.Var("namespace")

	metrics, err := Provider.MetricList(app, ns, types.MetricListOptions{})
	if err != nil {
		return err
	}

	return c.RenderJSON(metrics)
}
