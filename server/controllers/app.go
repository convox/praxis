package controllers

import (
	"fmt"
	"net/http"
)

func AppCreate(w http.ResponseWriter, r *http.Request) error {
	name := r.FormValue("name")
	fmt.Printf("name = %+v\n", name)

	w.Write([]byte(`{"name":"foo"}`))

	return nil
}
