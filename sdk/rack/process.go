package rack

import (
	"fmt"
	"io"
	"strconv"

	"github.com/convox/praxis/types"
)

func (c *Client) ProcessRun(app string, opts types.ProcessRunOptions) error {
	ro := RequestOptions{
		Body: opts.Stream,
		Headers: Headers{
			"Command": opts.Command,
			"Service": opts.Service,
		},
	}

	if opts.Height > 0 {
		ro.Headers["Height"] = strconv.Itoa(opts.Height)
	}

	if opts.Width > 0 {
		ro.Headers["Width"] = strconv.Itoa(opts.Width)
	}

	r, err := c.PostStream(fmt.Sprintf("/apps/%s/processes", app), ro)
	if err != nil {
		return err
	}

	if _, err := io.Copy(opts.Stream, r); err != nil {
		return err
	}

	return nil
}
