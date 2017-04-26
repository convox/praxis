package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) ResourceCreate(name, kind string, params map[string]string) (res *types.Resource, err error) {
	pa := Params{
		"name": name,
		"kind": kind,
	}

	for k, v := range params {
		pa[k] = v
	}

	ro := RequestOptions{
		Params: pa,
	}

	err = c.Post("/resources", ro, &res)
	return
}

func (c *Client) ResourceGet(name string) (res *types.Resource, err error) {
	err = c.Get(fmt.Sprintf("/resources/%s", name), RequestOptions{}, &res)
	return
}

func (c *Client) ResourceList() (ress types.Resources, err error) {
	err = c.Get("/resources", RequestOptions{}, &ress)
	return
}
