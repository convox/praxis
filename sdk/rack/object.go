package rack

import "io"

type Object struct {
	URL string
}

func (c *Client) ObjectStore(app string, r io.Reader) (*Object, error) {
	return &Object{URL: "http://example.org/test.tgz"}, nil
}
