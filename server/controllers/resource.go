package controllers

import (
	"net/http"

	"github.com/convox/api"
)

func ResourceCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	kind := c.Form("kind")
	name := c.Form("name")

	c.LogParams("name")

	resource, err := Provider.ResourceCreate(kind, name, nil)
	if err != nil {
		return err
	}

	return c.RenderJSON(resource)
}

func ResourceGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	name := c.Var("name")

	resource, err := Provider.ResourceGet(name)
	if err != nil {
		return err
	}

	return c.RenderJSON(resource)
}

func ResourceList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	ress, err := Provider.ResourceList()
	if err != nil {
		return err
	}

	return c.RenderJSON(ress)
}
