package controllers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/convox/praxis/provider"
	"github.com/convox/praxis/provider/local"
)

var (
	Provider = providerFromEnv()
)

func Render(w http.ResponseWriter, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func providerFromEnv() provider.Provider {
	switch os.Getenv("PROVIDER") {
	default:
		return local.FromEnv()
	}
}
