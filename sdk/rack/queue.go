package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) QueueFetch(app, queue string, opts types.QueueFetchOptions) (attrs map[string]string, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/queues/%s", app, queue), RequestOptions{}, &attrs)
	return
}

func (c *Client) QueueStore(app, queue string, attrs map[string]string) error {
	params := map[string]interface{}{}

	for k, v := range attrs {
		params[k] = v
	}

	ro := RequestOptions{
		Params: params,
	}

	return c.Post(fmt.Sprintf("/apps/%s/queues/%s", app, queue), ro, nil)
}
