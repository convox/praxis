package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) CacheFetch(app, cache, key string) (attrs map[string]string, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/caches/%s/%s", app, cache, key), RequestOptions{}, &attrs)
	return
}

func (c *Client) CacheStore(app, cache, key string, attrs map[string]string, opts types.CacheStoreOptions) error {
	params := map[string]interface{}{}

	for k, v := range attrs {
		params[k] = v
	}

	ro := RequestOptions{
		Params: params,
	}

	return c.Post(fmt.Sprintf("/apps/%s/caches/%s/%s", app, cache, key), ro, nil)
}
