package local

import (
	"fmt"
	"time"

	lcache "github.com/convox/praxis/cache"
	"github.com/convox/praxis/types"
)

func (p *Provider) CacheFetch(app, cache, key string) (map[string]string, error) {
	collection := fmt.Sprintf("%s-%s-%s", app, cache, key)

	item := lcache.Get(collection, key)
	if item != nil {
		if attrs, ok := item.(map[string]string); ok {
			return attrs, nil
		}

		return nil, fmt.Errorf("cache item not of type map[string]string")
	}

	return nil, nil
}

func (p *Provider) CacheStore(app, cache, key string, attrs map[string]string, opts types.CacheStoreOptions) error {
	collection := fmt.Sprintf("%s-%s-%s", app, cache, key)
	ttl := time.Duration(coalescei(opts.Expires, 60)) * time.Second

	return lcache.Set(collection, key, attrs, ttl)
}
