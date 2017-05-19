package local

import (
	"fmt"
	"time"

	lcache "github.com/convox/praxis/cache"
	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

func (p *Provider) CacheFetch(app, cache, key string) (map[string]string, error) {
	log := p.logger("CacheFetch").Append("app=%q cache=%q key=%q", app, cache, key)

	collection := fmt.Sprintf("%s-%s-%s", app, cache, key)

	item := lcache.Get(collection, key)
	if item != nil {
		if attrs, ok := item.(map[string]string); ok {
			return attrs, nil
		}

		return nil, log.Error(fmt.Errorf("cache item not of type map[string]string"))
	}

	return nil, log.Success()
}

func (p *Provider) CacheStore(app, cache, key string, attrs map[string]string, opts types.CacheStoreOptions) error {
	log := p.logger("CacheStore").Append("app=%q cache=%q key=%q", app, cache, key)

	collection := fmt.Sprintf("%s-%s-%s", app, cache, key)
	ttl := time.Duration(coalescei(opts.Expires, 60)) * time.Second

	if err := lcache.Set(collection, key, attrs, ttl); err != nil {
		return errors.WithStack(log.Error(err))
	}

	return log.Success()
}
