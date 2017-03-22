package controllers

import (
	"net/http"

	"github.com/convox/api"
)

func RegistryAdd(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	hostname := c.Form("hostname")
	username := c.Form("username")
	password := c.Form("password")

	registry, err := Provider.RegistryAdd(hostname, username, password)
	if err != nil {
		return err
	}

	return c.RenderJSON(registry)
}

func RegistryList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	registries, err := Provider.RegistryList()
	if err != nil {
		return err
	}

	return c.RenderJSON(registries)
}

func RegistryRemove(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	hostname := c.Var("hostname")

	return Provider.RegistryRemove(hostname)
}
