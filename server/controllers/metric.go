package controllers

import (
	"net/http"

	"github.com/convox/praxis/api"
)

func MetricGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	name := c.Var("name")

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

	metric, err := Provider.MetricGet(app, name)
	if err != nil {
		return err
	}

	return c.RenderJSON(metric)
}

func MetricList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

	metrics, err := Provider.MetricList(app)
	if err != nil {
		return err
	}

	return c.RenderJSON(metrics)
}
