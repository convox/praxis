package controllers

import (
	"io/ioutil"
	"net/http"

	"github.com/convox/api"
)

func KeyDecrypt(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	key := c.Var("key")

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	data, err := Provider.KeyDecrypt(app, key, bytes)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func KeyEncrypt(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	key := c.Var("key")

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	data, err := Provider.KeyEncrypt(app, key, bytes)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}
