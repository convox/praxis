package aws

import (
	"fmt"

	"github.com/convox/praxis/types"
)

var cacheBuffer = make(map[string]interface{})

func (p *Provider) CacheFetch(app, cache, key string) (map[string]string, error) {
	return nil, nil
}

func (p *Provider) CacheStore(app, cache, key string, attrs map[string]string, opts types.CacheStoreOptions) error {
	return fmt.Errorf("unimplemented")
}
