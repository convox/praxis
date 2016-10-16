package controllers

import (
	"fmt"
	"net/http"

	"github.com/convox/praxis/provider/models"
)

func BuildCreate(w http.ResponseWriter, r *http.Request) error {
	fmt.Println(w)
	build, err := Provider.BuildCreate(r.FormValue("url"), models.BuildCreateOptions{
		Cache: true,
	})
	fmt.Printf("build = %+v\n", build)
	fmt.Printf("err = %+v\n", err)
	if err != nil {
		return err
	}

	return Render(w, build)
}
