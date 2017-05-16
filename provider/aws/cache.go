package aws

import (
	"fmt"
	"time"

	lcache "github.com/convox/praxis/cache"
	"github.com/convox/praxis/types"
	"github.com/go-redis/redis"
)

func (p *Provider) CacheFetch(app, cache, key string) (map[string]string, error) {
	client, err := p.redisClient(app, cache)
	if err != nil {
		return nil, err
	}

	return client.HGetAll(fmt.Sprintf("%s:%s:%s", app, cache, key)).Result()
}

func (p *Provider) CacheStore(app, cache, key string, attrs map[string]string, opts types.CacheStoreOptions) error {
	client, err := p.redisClient(app, cache)
	if err != nil {
		return err
	}

	data := map[string]interface{}{}
	for k, v := range attrs {
		data[k] = v
	}

	return client.HMSet(fmt.Sprintf("%s:%s:%s", app, cache, key), data).Err()
}

func (p *Provider) redisClient(app, cache string) (*redis.Client, error) {
	ep := lcache.Get("redis_endpoints", fmt.Sprintf("%s:%s", app, cache))
	if ep == nil {
		re, err := p.appOutput(app, fmt.Sprintf("Cache%sEndpoint", upperName(cache)))
		if err != nil {
			return nil, fmt.Errorf("redis endpoint: %s", err)
		}

		rp, err := p.appOutput(app, fmt.Sprintf("Cache%sPort", upperName(cache)))
		if err != nil {
			return nil, fmt.Errorf("redis port: %s", err)
		}

		ep = fmt.Sprintf("%s:%s", re, rp)

		err = lcache.Set("redis_endpoints", fmt.Sprintf("%s:%s", app, cache), ep, 1*time.Hour)
		if err != nil {
			fmt.Printf("redis endpoint not cached: %s", ep)
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     ep.(string),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("redis pong: %s", err)
	}

	return client, nil
}
