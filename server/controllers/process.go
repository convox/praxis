package controllers

import (
	"net/http"
	"strconv"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/types"
)

func ProcessRun(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	service := c.Header("Service")
	command := c.Header("Command")
	height := c.Header("Height")
	width := c.Header("Width")

	opts := types.ProcessRunOptions{
		Command: command,
		Service: service,
		Stream: types.Stream{
			Reader: r.Body,
			Writer: w,
		},
	}

	if height != "" {
		h, err := strconv.Atoi(height)
		if err != nil {
			return err
		}

		opts.Height = h
	}

	if width != "" {
		w, err := strconv.Atoi(width)
		if err != nil {
			return err
		}

		opts.Width = w
	}

	w.Header().Add("Trailer", "Exit-Code")

	if err := Provider.ProcessRun(app, opts); err != nil {
		return err
	}

	w.Header().Set("Exit-Code", "2")

	return nil
}
