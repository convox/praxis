package controllers

import (
	"net/http"
	"sort"

	"github.com/convox/praxis/api"
)

func ResourceList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	_, err := Provider.AppGet(app)
	if err != nil {
		return err
	}

	ss, err := Provider.ResourceList(app)
	if err != nil {
		return err
	}

	sort.Slice(ss, func(i, j int) bool { return ss[i].Name < ss[j].Name })

	return c.RenderJSON(ss)
}
