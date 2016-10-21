package controllers

import (
	"net/http"

	"github.com/convox/praxis/provider"
	"github.com/gorilla/mux"
)

func AppCreate(w http.ResponseWriter, r *http.Request) error {
	app, err := Provider.AppCreate(r.FormValue("name"), provider.AppCreateOptions{})
	if err != nil {
		return err
	}

	return Render(w, app)
}

func AppDelete(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	if err := Provider.AppDelete(vars["app"]); err != nil {
		return err
	}

	return Render(w, "ok")
}
