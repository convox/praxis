package controllers

import (
	"fmt"
	"net/http"

	"github.com/convox/praxis/api"
)

func AppCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	name := c.Form("name")

	c.LogParams("name")

	app, err := Provider.AppCreate(name)
	if err != nil {
		return err
	}

	return c.RenderJSON(app)
}

func AppDelete(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	fmt.Printf("app = %+v\n", app)

	if err := Provider.AppDelete(app); err != nil {
		return err
	}

	return nil
}
