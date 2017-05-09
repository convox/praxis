package local

import (
	"fmt"
	"time"

	lcache "github.com/convox/praxis/cache"
	"github.com/convox/praxis/types"
)

var cacheBuffer = make(map[string]interface{})

func (p *Provider) CacheFetch(app, cache, key string) map[string]string {
	collection := fmt.Sprintf("%s-%s-%s", app, cache, key)

	attrs, ok := lcache.Get(collection, key).(map[string]string)
	if !ok {
		return nil
	}

	return attrs
}

func (p *Provider) CacheStore(app, cache, key string, attrs map[string]string, opts types.CacheStoreOptions) error {
	collection := fmt.Sprintf("%s-%s-%s", app, cache, key)
	ttl := time.Duration(coalescei(opts.Expires, 60)) * time.Second

	return lcache.Set(collection, key, attrs, ttl)
}
