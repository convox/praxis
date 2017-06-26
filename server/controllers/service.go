package controllers

import (
	"net/http"
	"sort"

	"github.com/convox/praxis/api"
)

func ServiceGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	name := c.Var("name")

	s, err := Provider.ServiceGet(app, name)
	if err != nil {
		return err
	}

	return c.RenderJSON(s)
}

func ServiceList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	_, err := Provider.AppGet(app)
	if err != nil {
		return err
	}

	ss, err := Provider.ServiceList(app)
	if err != nil {
		return err
	}

	sort.Slice(ss, func(i, j int) bool { return ss[i].Name < ss[j].Name })

	return c.RenderJSON(ss)
}
