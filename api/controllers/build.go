package controllers

import (
	"io"
	"net/http"

	"github.com/convox/praxis/provider"
	"github.com/gorilla/mux"
)

func BuildCreate(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	build, err := Provider.BuildCreate(vars["app"], r.FormValue("url"), provider.BuildCreateOptions{
		Cache: true,
	})
	if err != nil {
		return err
	}

	return Render(w, build)
}

func BuildGet(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	build, err := Provider.BuildLoad(vars["app"], vars["build"])
	if err != nil {
		return err
	}

	return Render(w, build)
}

func BuildLogs(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	rd, err := Provider.BuildLogs(vars["app"], vars["build"])
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, rd); err != nil {
		return err
	}

	return nil
}
