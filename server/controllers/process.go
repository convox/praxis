package controllers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/types"
)

func ProcessExec(rw io.ReadWriteCloser, c *api.Context) error {
	app := c.Var("app")
	pid := c.Var("pid")

	command := c.Header("Command")
	height := c.Header("Height")
	width := c.Header("Width")

	opts := types.ProcessExecOptions{
		Stream: rw,
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

	// if r.ProtoAtLeast(2, 0) {
	//   w.Header().Add("Trailer", "Exit-Code")
	// }

	code, err := Provider.ProcessExec(app, pid, command, opts)
	if err != nil {
		return err
	}

	fmt.Printf("code = %+v\n", code)

	// if r.ProtoAtLeast(2, 0) {
	//   w.Header().Set("Exit-Code", fmt.Sprintf("%d", code))
	// } else {
	//   fmt.Fprintf(opts.Stream, "SOMEUUID: %d\n", code)
	//   fmt.Fprintf(w, "SOMEUUID: %d\n", code)
	// }

	return err
}

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
	service := c.Query("service")

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

	opts := types.LogsOptions{
		Follow: c.Query("follow") == "true",
		Prefix: c.Query("prefix") == "true",
	}

	logs, err := Provider.ProcessLogs(app, pid, opts)
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

	command := c.Header("Command")
	height := c.Header("Height")
	image := c.Header("Image")
	links := c.Header("Links")
	name := c.Header("Name")
	release := c.Header("Release")
	service := c.Header("Service")
	width := c.Header("Width")

	env := map[string]string{}

	ev, err := url.ParseQuery(c.Header("Environment"))
	if err != nil {
		return err
	}

	for k := range ev {
		env[k] = ev.Get(k)
	}

	ports := map[int]int{}

	pv, err := url.ParseQuery(c.Header("Ports"))
	if err != nil {
		return err
	}

	for k := range pv {
		ki, err := strconv.Atoi(k)
		if err != nil {
			return err
		}

		vi, err := strconv.Atoi(pv.Get(k))
		if err != nil {
			return err
		}

		ports[ki] = vi
	}

	volumes := map[string]string{}

	vv, err := url.ParseQuery(c.Header("Volumes"))
	if err != nil {
		return err
	}

	for k := range vv {
		volumes[k] = vv.Get(k)
	}

	opts := types.ProcessRunOptions{
		Command:     command,
		Environment: env,
		Image:       image,
		Name:        name,
		Ports:       ports,
		Release:     release,
		Service:     service,
		Stream:      types.Stream{Reader: r.Body, Writer: w},
		Volumes:     volumes,
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

	if links != "" {
		opts.Links = strings.Split(links, ",")
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
	image := c.Form("image")
	links := c.Form("links")
	name := c.Form("name")
	release := c.Form("release")
	service := c.Form("service")

	env := map[string]string{}

	ev, err := url.ParseQuery(c.Form("environment"))
	if err != nil {
		return err
	}

	for k := range ev {
		env[k] = ev.Get(k)
	}

	ports := map[int]int{}

	pv, err := url.ParseQuery(c.Form("ports"))
	if err != nil {
		return err
	}

	for k := range pv {
		ki, err := strconv.Atoi(k)
		if err != nil {
			return err
		}

		vi, err := strconv.Atoi(pv.Get(k))
		if err != nil {
			return err
		}

		ports[ki] = vi
	}

	volumes := map[string]string{}

	vv, err := url.ParseQuery(c.Form("volumes"))
	if err != nil {
		return err
	}

	for k := range vv {
		volumes[k] = vv.Get(k)
	}

	opts := types.ProcessRunOptions{
		Command:     command,
		Environment: env,
		Image:       image,
		Name:        name,
		Ports:       ports,
		Release:     release,
		Service:     service,
		Volumes:     volumes,
	}

	if links != "" {
		opts.Links = strings.Split(links, ",")
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
