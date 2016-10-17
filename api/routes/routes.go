package routes

import (
	"net/http"

	"github.com/convox/praxis/api/controllers"
	"github.com/gorilla/mux"
)

func New() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/apps", api(controllers.AppCreate)).Methods("POST")
	r.HandleFunc("/apps/{app}/builds", api(controllers.BuildCreate)).Methods("POST")
	r.HandleFunc("/apps/{app}/blobs/{key:.*}", api(controllers.BlobStore)).Methods("POST")

	return r
}

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func api(fn handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}
