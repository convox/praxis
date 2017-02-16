package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) ReleaseCreate(app string, opts types.ReleaseCreateOptions) (release *types.Release, err error) {
	ro := RequestOptions{
		Params: Params{
			"build": opts.Build,
			"env":   fmt.Sprintf("%v", opts.Env),
		},
	}

	err = c.Post(fmt.Sprintf("/apps/%s/releases", app), ro, &release)

	return
}

func (c *Client) ReleaseGet(app, id string) (release *types.Release, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/releases/%s", app, id), RequestOptions{}, &release)
	return
}
