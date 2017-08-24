package rack

import (
	"fmt"
	"io"

	"github.com/convox/praxis/types"
)

func (c *Client) ResourceGet(app, name string) (r *types.Resource, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/resources/%s", app, name), RequestOptions{}, &r)
	return
}

func (c *Client) ResourceList(app string) (rs types.Resources, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/resources", app), RequestOptions{}, &rs)
	return
}

func (c *Client) ResourceProxy(app, resource string, in io.Reader) (io.ReadCloser, error) {
	ro := RequestOptions{
		Body: in,
	}

	r, err := c.Stream(fmt.Sprintf("/apps/%s/resources/%s/proxy", app, resource), ro)
	if err != nil {
		return nil, err
	}

	return r, nil
}
