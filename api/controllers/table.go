package controllers

import (
	"fmt"
	"net/http"
)

func TableList(w http.ResponseWriter, r *http.Request) {
	tables, err := Provider.TableList()
	fmt.Printf("tables = %+v\n", tables)
	fmt.Printf("err = %+v\n", err)
	if err != nil {
		RenderError(w, err)
		return
	}

	RenderJSON(w, tables)
}
