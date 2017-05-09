package controllers

import (
	"fmt"
	"net/http"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/types"
)

func CacheFetch(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	cache := c.Var("cache")
	key := c.Var("key")

	attrs, err := Provider.CacheFetch(app, cache, key)
	if err != nil {
		return err
	}

	return c.RenderJSON(attrs)
}

func CacheStore(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	cache := c.Var("cache")
	key := c.Var("key")

	if err := r.ParseForm(); err != nil {
		return err
	}

	attrs := map[string]string{}

	for k := range r.Form {
		attrs[k] = r.Form.Get(k)
	}

	err := Provider.CacheStore(app, cache, key, attrs, types.CacheStoreOptions{})
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "ok")

	return nil
}
