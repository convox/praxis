package controllers

import (
	"net/http"

	"github.com/convox/api"
	"github.com/convox/praxis/types"
)

func EnvironmentDelete(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	key := c.Var("key")

	return Provider.EnvironmentDelete(app, key)
}

func EnvironmentGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	env, err := Provider.EnvironmentGet(app)
	if err != nil {
		return err
	}

	return c.RenderJSON(env)
}

func EnvironmentSet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	if err := r.ParseForm(); err != nil {
		return err
	}

	env := types.Environment{}

	for k := range r.Form {
		env[k] = r.Form.Get(k)
	}

	if err := Provider.EnvironmentSet(app, env); err != nil {
		return err
	}

	e, err := Provider.EnvironmentGet(app)
	if err != nil {
		return err
	}

	rl, err := Provider.ReleaseCreate(app, types.ReleaseCreateOptions{Env: e})
	if err != nil {
		return err
	}

	w.Header().Add("Release", rl.Id)

	return nil
}
