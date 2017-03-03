package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/convox/praxis/api"
)

func FilesDelete(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	pid := c.Var("process")

	if strings.TrimSpace(pid) == "" {
		return fmt.Errorf("must specify a pid")
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	uv, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	files := strings.Split(uv.Get("files"), ",")

	if len(files) == 0 {
		return fmt.Errorf("must specify at least one file")
	}

	if err := Provider.FilesDelete(app, pid, files); err != nil {
		return err
	}

	return nil
}

func FilesUpload(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	pid := c.Var("process")

	return Provider.FilesUpload(app, pid, r.Body)
}
