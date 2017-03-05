package controllers

import (
	"net/http"
	"strconv"

	"github.com/convox/api"
)

func ProxyStart(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	process := c.Var("process")
	port := c.Var("port")

	pi, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	proxy, err := Provider.ProxyStart(app, process, pi)
	if err != nil {
		return err
	}

	w.WriteHeader(200)

	go stream(proxy, r.Body)

	if err := stream(w, proxy); err != nil {
		return err
	}

	return nil
}
