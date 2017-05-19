package controllers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

func ObjectExists(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	key := c.Var("key")

	exists, err := Provider.ObjectExists(app, key)
	if err != nil {
		return err
	}

	status := http.StatusNotFound
	if exists {
		status = http.StatusOK
	}

	w.WriteHeader(status)
	return nil
}

func ObjectFetch(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	key := c.Var("key")

	obj, err := Provider.ObjectFetch(app, key)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, obj); err != nil {
		return err
	}

	return nil
}

func ObjectStore(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	key := c.Var("key")

	if key == "" {
		k, err := types.Key(32)
		if err != nil {
			return errors.WithStack(err)
		}

		key = fmt.Sprintf("tmp/%s", k)
	}

	object, err := Provider.ObjectStore(app, key, r.Body, types.ObjectStoreOptions{})
	if err != nil {
		return err
	}

	return c.RenderJSON(object)
}
