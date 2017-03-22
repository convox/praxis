package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) EnvironmentGet(app string) (env types.Environment, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/environment", app), RequestOptions{}, &env)
	return
}

func (c *Client) EnvironmentSet(app string, env types.Environment) error {
	params := Params{}

	for k, v := range env {
		params[k] = v
	}

	ro := RequestOptions{
		Params: params,
	}

	return c.Post(fmt.Sprintf("/apps/%s/environment", app), ro, nil)
}

func (c *Client) EnvironmentUnset(app string, key string) error {
	return c.Delete(fmt.Sprintf("/apps/%s/environment/%s", app, key), RequestOptions{}, nil)
}
