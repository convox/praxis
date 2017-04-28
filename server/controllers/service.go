package controllers

import (
	"net/http"
	"sort"

	"github.com/convox/api"
)

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
