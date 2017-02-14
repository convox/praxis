package controllers

import (
	"net/http"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/provider/types"
)

func ObjectStore(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	key := c.Var("key")

	object, err := Provider.ObjectStore(app, key, r.Body, types.ObjectStoreOptions{})
	if err != nil {
		return err
	}

	return c.RenderJSON(object)
}
