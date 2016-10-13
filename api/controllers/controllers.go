package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/convox/praxis/provider"
)

var (
	Provider = provider.FromEnv()
)

func RenderError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), 500)
}

func RenderJSON(w http.ResponseWriter, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		RenderError(w, err)
		return
	}

	w.Write(data)
}
