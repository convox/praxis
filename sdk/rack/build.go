package rack

import (
	"fmt"

	"github.com/convox/praxis/provider/types"
)

type Build types.Build

func (c *Client) BuildCreate(app string, url string) (build *Build, err error) {
	err = c.Post(fmt.Sprintf("/apps/%s/builds", app), Params{"url": url}, &build)
	return
}
