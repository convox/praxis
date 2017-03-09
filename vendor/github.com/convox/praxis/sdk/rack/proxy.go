package rack

import (
	"fmt"
	"io"
)

func (c *Client) Proxy(app, pid string, port int, in io.Reader) (io.ReadCloser, error) {
	ro := RequestOptions{
		Body: in,
	}

	res, err := c.PostStream(fmt.Sprintf("/apps/%s/processes/%s/proxy/%d", app, pid, port), ro)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
