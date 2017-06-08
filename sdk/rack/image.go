package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) ImageCreate(name, url string, opts types.ImageCreateOptions) (*types.Image, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (c *Client) ImageList() (types.Images, error) {
	return nil, fmt.Errorf("unimplemented")
}
