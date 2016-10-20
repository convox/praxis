package controllers

import (
	"net/http"

	"github.com/convox/praxis/provider/models"
	"github.com/gorilla/mux"
)

func BuildCreate(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	build, err := Provider.BuildCreate(vars["app"], r.FormValue("url"), models.BuildCreateOptions{
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
