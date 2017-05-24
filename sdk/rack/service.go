package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) ServiceGet(app, name string) (s *types.Service, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/services/%s", app, name), RequestOptions{}, &s)
	return
}

func (c *Client) ServiceList(app string) (ss types.Services, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/services", app), RequestOptions{}, &ss)
	return
}
