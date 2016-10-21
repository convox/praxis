package controllers

import (
	"io"
	"net/http"

	"github.com/convox/praxis/provider"
	"github.com/gorilla/mux"
)

func BlobFetch(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	fd, err := Provider.BlobFetch(vars["app"], vars["key"])
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, fd); err != nil {
		return err
	}

	return nil
}

func BlobStore(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	data, _, err := r.FormFile("data")
	if err != nil {
		return err
	}

	url, err := Provider.BlobStore(vars["app"], vars["key"], data, provider.BlobStoreOptions{
		Public: r.FormValue("public") == "true",
	})
	if err != nil {
		return err
	}

	return Render(w, url)
}
