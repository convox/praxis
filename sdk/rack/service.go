package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) ServiceList(app string) (ss types.Services, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/services", app), RequestOptions{}, &ss)
	return
}
