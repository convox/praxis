package controllers

import (
	"fmt"
	"net/http"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/provider/types"
)

func ObjectStore(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	key := c.Var("key")

	if key == "" {
		r, err := randomString()
		if err != nil {
			return err
		}

		key = fmt.Sprintf("tmp/%s", r)
	}

	object, err := Provider.ObjectStore(app, key, r.Body, types.ObjectStoreOptions{})
	if err != nil {
		return err
	}

	return c.RenderJSON(object)
}
