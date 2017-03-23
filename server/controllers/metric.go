package controllers

import (
	"net/http"

	"github.com/convox/api"
	"github.com/convox/praxis/types"
)

func MetricList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	ns := c.Var("namespace")

	metrics, err := Provider.MetricList(app, ns, types.MetricListOptions{})
	if err != nil {
		return err
	}

	return c.RenderJSON(metrics)
}
