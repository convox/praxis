package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) ResourceList(app string) (rs types.Resources, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/resources", app), RequestOptions{}, &rs)
	return
}
