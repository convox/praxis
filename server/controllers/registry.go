package controllers

import (
	"net/http"

	"github.com/convox/api"
)

func RegistryAdd(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	server := c.Form("server")
	username := c.Form("username")
	password := c.Form("password")

	registry, err := Provider.RegistryAdd(server, username, password)
	if err != nil {
		return err
	}

	return c.RenderJSON(registry)
}

func RegistryDelete(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	server := c.Form("server")

	return Provider.RegistryDelete(server)
}

func RegistryList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	registries, err := Provider.RegistryList()
	if err != nil {
		return err
	}

	return c.RenderJSON(registries)
}
