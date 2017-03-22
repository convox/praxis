package rack

import "github.com/convox/praxis/types"

func (c *Client) SystemGet() (system *types.System, err error) {
	err = c.Get("/system", RequestOptions{}, &system)
	return
}

func (c *Client) SystemImage() (image string, err error) {
	err = c.Get("/system/image", RequestOptions{}, &image)
	return
}
