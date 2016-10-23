package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/convox/praxis/provider"
	"github.com/convox/praxis/provider/local"
)

var (
	Provider = providerFromEnv()
)

func providerFromEnv() provider.Provider {
	switch os.Getenv("PROVIDER") {
	default:
		return local.FromEnv()
	}
}

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

type flushWriter struct {
	w io.Writer
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)

	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}

	return
}
