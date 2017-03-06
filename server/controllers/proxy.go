package controllers

import (
	"net/http"
	"strconv"

	"github.com/convox/api"
)

func Proxy(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	process := c.Var("process")
	port := c.Var("port")

	pi, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	out, err := Provider.Proxy(app, process, pi, r.Body)
	if err != nil {
		return err
	}

	defer out.Close()

	w.WriteHeader(200)

	if err := stream(w, out); err != nil {
		return err
	}

	return nil
}
