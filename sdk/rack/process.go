package rack

import (
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/convox/praxis/types"
)

func (c *Client) ProcessList(app string, opts types.ProcessListOptions) (ps types.Processes, err error) {
	ro := RequestOptions{
		Params: Params{
			"service": opts.Service,
		},
	}

	err = c.Get(fmt.Sprintf("/apps/%s/processes", app), ro, &ps)
	return
}

func (c *Client) ProcessRun(app string, opts types.ProcessRunOptions) (int, error) {
	ev := url.Values{}

	for k, v := range opts.Environment {
		ev.Add(k, v)
	}

	ro := RequestOptions{
		Body: opts.Stream,
		Headers: Headers{
			"Command":     opts.Command,
			"Environment": ev.Encode(),
			"Release":     opts.Release,
			"Service":     opts.Service,
		},
	}

	if opts.Height > 0 {
		ro.Headers["Height"] = strconv.Itoa(opts.Height)
	}

	if opts.Width > 0 {
		ro.Headers["Width"] = strconv.Itoa(opts.Width)
	}

	res, err := c.PostStream(fmt.Sprintf("/apps/%s/processes", app), ro)
	if err != nil {
		return 0, err
	}

	defer res.Body.Close()

	if _, err := io.Copy(opts.Stream, res.Body); err != nil {
		return 0, err
	}

	if code, err := strconv.Atoi(res.Trailer.Get("Exit-Code")); err == nil {
		return code, nil
	}

	return 0, nil
}

func (c *Client) ProcessStop(app, pid string) error {
	return c.Delete(fmt.Sprintf("/apps/%s/processes/%s", app, pid), RequestOptions{}, nil)
}
