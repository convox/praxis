package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) AppCreate(name string) (app *types.App, err error) {
	ro := RequestOptions{
		Params: Params{
			"name": name,
		},
	}

	err = c.Post("/apps", ro, &app)
	return
}

func (c *Client) AppDelete(name string) error {
	return c.Delete(fmt.Sprintf("/apps/%s", name), RequestOptions{}, nil)
}

func (c *Client) AppList() (apps types.Apps, err error) {
	err = c.Get("/apps", RequestOptions{}, &apps)
	return
}
