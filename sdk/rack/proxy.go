package rack

import (
	"fmt"
	"io"
)

func (c *Client) ProxyStart(app, pid string, port int, stream io.ReadWriter) error {
	ro := RequestOptions{
		Body: stream,
	}

	res, err := c.PostStream(fmt.Sprintf("/apps/%s/processes/%s/proxy/%d", app, pid, port), ro)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if _, err := io.Copy(stream, res.Body); err != nil {
		return err
	}

	return nil
}
