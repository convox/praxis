package controllers

import (
	"fmt"
	"net/http"

	"github.com/convox/praxis/api"
)

func AppCreate(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	// name := r.FormValue("name")
	// fmt.Printf("name = %+v\n", name)

	// if name == "" {
	//   return api.Errorf(400, "name must not be blank")
	// }

	c.LogParams("name")

	w.Write([]byte(`{"name":"foo"}`))
	return nil

	return fmt.Errorf("test")
}
