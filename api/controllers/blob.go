package controllers

import (
	"net/http"

	"github.com/convox/praxis/provider/models"
	"github.com/gorilla/mux"
)

func BlobStore(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	data, _, err := r.FormFile("data")
	if err != nil {
		return err
	}

	url, err := Provider.BlobStore(vars["app"], vars["key"], data, models.BlobStoreOptions{
		Public: r.FormValue("public") == "true",
	})
	if err != nil {
		return err
	}

	return Render(w, url)
}
