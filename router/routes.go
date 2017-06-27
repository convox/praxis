package router

import (
	"fmt"
	"net/http"

	"github.com/convox/praxis/api"
)

func (rt *Router) EndpointCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	host := c.Var("host")

	ep, err := rt.createEndpoint(host)
	if err != nil {
		return err
	}

	return c.RenderJSON(ep)
}

func (rt *Router) EndpointDelete(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	return nil
}

func (rt *Router) EndpointList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	return c.RenderJSON(rt.endpoints)
}

func (rt *Router) ProxyCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	host := c.Var("host")
	port := c.Var("port")
	scheme := c.Form("scheme")
	target := c.Form("target")

	ep, ok := rt.endpoints[host]
	if !ok {
		return fmt.Errorf("no such endpoint: %s", host)
	}

	if _, err := rt.createProxy(host, fmt.Sprintf("%s://%s:%s", scheme, ep.IP, port), target); err != nil {
		return err
	}

	return nil
}
