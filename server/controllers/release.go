package controllers

import (
	"net/http"
	"sort"

	"github.com/convox/api"
	"github.com/convox/praxis/types"
)

func ReleaseCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	build := c.Form("build")
	env := map[string]string{}

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

	release, err := Provider.ReleaseCreate(app, types.ReleaseCreateOptions{
		Build: build,
		Env:   env,
	})
	if err != nil {
		return err
	}

	return c.RenderJSON(release)
}

func ReleaseGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	id := c.Var("id")

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

	release, err := Provider.ReleaseGet(app, id)
	if err != nil {
		return err
	}

	return c.RenderJSON(release)
}

func ReleaseList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

	releases, err := Provider.ReleaseList(app)
	if err != nil {
		return err
	}

	sort.Slice(releases, func(i, j int) bool { return releases[j].Created.Before(releases[i].Created) })

	if len(releases) > 10 {
		releases = releases[0:10]
	}

	return c.RenderJSON(releases)
}

func ReleaseLogs(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	id := c.Var("id")

	if _, err := Provider.AppGet(app); err != nil {
		return err
	}

	logs, err := Provider.ReleaseLogs(app, id)
	if err != nil {
		return err
	}

	if err := stream(w, logs); err != nil {
		return err
	}

	return nil
}
