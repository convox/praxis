package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/convox/api"
	"github.com/convox/praxis/types"
)

func ProcessGet(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	pid := c.Var("pid")

	ps, err := Provider.ProcessGet(app, pid)
	if err != nil {
		return err
	}

	return c.RenderJSON(ps)
}

func ProcessList(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	service := c.Form("service")

	c.LogParams("service")

	_, err := Provider.AppGet(app)
	if err != nil {
		return err
	}

	opts := types.ProcessListOptions{
		Service: service,
	}

	ps, err := Provider.ProcessList(app, opts)
	if err != nil {
		return err
	}

	return c.RenderJSON(ps)
}

func ProcessLogs(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	pid := c.Var("pid")

	logs, err := Provider.ProcessLogs(app, pid)
	if err != nil {
		return err
	}

	if err := stream(w, logs); err != nil {
		return err
	}

	return nil
}

func ProcessRun(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	release := c.Header("Release")
	service := c.Header("Service")
	command := c.Header("Command")
	height := c.Header("Height")
	width := c.Header("Width")

	uenv, err := url.ParseQuery(c.Header("Environment"))
	if err != nil {
		return err
	}

	env := map[string]string{}

	for k := range uenv {
		env[k] = uenv.Get(k)
	}

	opts := types.ProcessRunOptions{
		Command:     command,
		Environment: env,
		Release:     release,
		Service:     service,
		Stream:      types.Stream{Reader: r.Body, Writer: w},
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

	if opts.Release == "" {
		a, err := Provider.AppGet(app)
		if err != nil {
			return err
		}

		if a.Release == "" {
			return fmt.Errorf("no releases for app: %s", app)
		}

		opts.Release = a.Release
	}

	c.Logf("at=params release=%q service=%q height=%d width=%d", opts.Release, opts.Service, opts.Height, opts.Width)

	w.Header().Add("Trailer", "Exit-Code")

	code, err := Provider.ProcessRun(app, opts)

	w.Header().Set("Exit-Code", fmt.Sprintf("%d", code))

	return err
}

func ProcessStart(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	command := c.Form("command")
	release := c.Form("release")
	service := c.Form("service")

	uenv, err := url.ParseQuery(c.Form("environment"))
	if err != nil {
		return err
	}

	env := map[string]string{}

	for k := range uenv {
		env[k] = uenv.Get(k)
	}

	opts := types.ProcessStartOptions{
		Command:     command,
		Environment: env,
		Release:     release,
		Service:     service,
	}

	if opts.Release == "" {
		a, err := Provider.AppGet(app)
		if err != nil {
			return err
		}

		if a.Release == "" {
			return fmt.Errorf("no releases for app: %s", app)
		}

		opts.Release = a.Release
	}

	c.LogParams("release", "service")

	pid, err := Provider.ProcessStart(app, opts)
	if err != nil {
		return err
	}

	return c.RenderJSON(pid)
}

func ProcessStop(w http.ResponseWriter, r *http.Request, c *api.Context) error {
	app := c.Var("app")
	pid := c.Var("pid")

	if err := Provider.ProcessStop(app, pid); err != nil {
		return err
	}

	return nil
}
