package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) CacheFetch(app, cache, key string) map[string]string {
	return nil
}

func (c *Client) CacheStore(app, cache, key string, attrs map[string]string, opts types.CacheStoreOptions) error {
	return fmt.Errorf("unimplemented")
}
