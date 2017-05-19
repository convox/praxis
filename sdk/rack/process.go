package rack

import (
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
)

func (c *Client) ProcessExec(app, pid, command string, opts types.ProcessExecOptions) (int, error) {
	ro := RequestOptions{
		Body: opts.Input,
		Headers: Headers{
			"Command": command,
		},
	}

	r, err := c.Stream(fmt.Sprintf("/apps/%s/processes/%s/exec", app, pid), ro)
	if err != nil {
		return 0, err
	}

	defer r.Close()

	if err := helpers.Stream(opts.Output, r); err != nil {
		return 0, err
	}

	// if code, err := strconv.Atoi(res.Trailer.Get("Exit-Code")); err == nil {
	//   return code, nil
	// }

	return 0, nil
}

func (c *Client) ProcessGet(app, pid string) (ps *types.Process, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/processes/%s", app, pid), RequestOptions{}, &ps)
	return
}

func (c *Client) ProcessList(app string, opts types.ProcessListOptions) (ps types.Processes, err error) {
	ro := RequestOptions{
		Query: Query{
			"service": opts.Service,
		},
	}

	err = c.Get(fmt.Sprintf("/apps/%s/processes", app), ro, &ps)
	return
}

func (c *Client) ProcessLogs(app, pid string, opts types.LogsOptions) (io.ReadCloser, error) {
	res, err := c.GetStream(fmt.Sprintf("/apps/%s/processes/%s/logs", app, pid), RequestOptions{})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) ProcessRun(app string, opts types.ProcessRunOptions) (int, error) {
	ev := url.Values{}

	for k, v := range opts.Environment {
		ev.Add(k, v)
	}

	pv := url.Values{}

	for k, v := range opts.Ports {
		pv.Add(fmt.Sprintf("%d", k), fmt.Sprintf("%d", v))
	}

	vv := url.Values{}

	for k, v := range opts.Volumes {
		vv.Add(k, v)
	}

	ro := RequestOptions{
		Body: opts.Input,
		Headers: Headers{
			"Command":     opts.Command,
			"Environment": ev.Encode(),
			"Image":       opts.Image,
			"Links":       strings.Join(opts.Links, ","),
			"Name":        opts.Name,
			"Ports":       pv.Encode(),
			"Release":     opts.Release,
			"Service":     opts.Service,
			"Volumes":     vv.Encode(),
		},
	}

	if opts.Height > 0 {
		ro.Headers["Height"] = strconv.Itoa(opts.Height)
	}

	if opts.Width > 0 {
		ro.Headers["Width"] = strconv.Itoa(opts.Width)
	}

	r, err := c.Stream(fmt.Sprintf("/apps/%s/processes/run", app), ro)
	if err != nil {
		return 0, err
	}

	defer r.Close()

	if err := helpers.Stream(opts.Output, r); err != nil {
		return 0, err
	}

	// TODO: get exit code

	return 0, nil
}

func (c *Client) ProcessStart(app string, opts types.ProcessRunOptions) (string, error) {
	ev := url.Values{}

	for k, v := range opts.Environment {
		ev.Add(k, v)
	}

	pv := url.Values{}

	for k, v := range opts.Ports {
		pv.Add(fmt.Sprintf("%d", k), fmt.Sprintf("%d", v))
	}

	vv := url.Values{}

	for k, v := range opts.Volumes {
		vv.Add(k, v)
	}

	ro := RequestOptions{
		Params: Params{
			"command":     opts.Command,
			"environment": ev.Encode(),
			"image":       opts.Image,
			"links":       strings.Join(opts.Links, ","),
			"name":        opts.Name,
			"ports":       pv.Encode(),
			"release":     opts.Release,
			"service":     opts.Service,
			"volumes":     vv.Encode(),
		},
	}

	var pid string

	if err := c.Post(fmt.Sprintf("/apps/%s/processes", app), ro, &pid); err != nil {
		return "", err
	}

	return pid, nil
}

func (c *Client) ProcessStop(app, pid string) error {
	return c.Delete(fmt.Sprintf("/apps/%s/processes/%s", app, pid), RequestOptions{}, nil)
}
