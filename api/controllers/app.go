package controllers

import (
	"net/http"

	"github.com/convox/praxis/provider/models"
)

func AppCreate(w http.ResponseWriter, r *http.Request) error {
	app, err := Provider.AppCreate(r.FormValue("name"), models.AppCreateOptions{})
	if err != nil {
		return err
	}

	return Render(w, app)
}
