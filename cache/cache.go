package cache

import (
	"strings"
	"sync"
	"time"
)

type Cache map[string]map[string]*CacheItem

type CacheItem struct {
	Item    interface{}
	Expires time.Time
}

var (
	cache = Cache{}
	lock  = sync.Mutex{}
)

func init() {
	go func() {
		for range time.Tick(1 * time.Minute) {
			Prune()
		}
	}()
}

func Get(collection string, key string) interface{} {
	lock.Lock()
	defer lock.Unlock()

	hash, err := hashKey(key)

	if err != nil {
		return nil
	}

	if cache[collection] == nil {
		return nil
	}

	item := cache[collection][hash]

	if item == nil {
		return nil
	}

	if item.Expires.Before(time.Now()) {
		return nil
	}

	return item.Item
}

func Set(collection string, key string, value interface{}, ttl time.Duration) error {
	lock.Lock()
	defer lock.Unlock()

	if cache[collection] == nil {
		cache[collection] = map[string]*CacheItem{}
	}

	hash, err := hashKey(key)

	if err != nil {
		return err
	}

	cache[collection][hash] = &CacheItem{
		Item:    value,
		Expires: time.Now().Add(ttl),
	}

	return nil
}

func Clear(collection string, key string) error {
	lock.Lock()
	defer lock.Unlock()

	hash, err := hashKey(key)

	if err != nil {
		return err
	}

	if cache[collection] != nil && cache[collection][hash] != nil {
		delete(cache[collection], hash)
	}

	return nil
}

func ClearPrefix(collection string, prefix string) error {
	lock.Lock()
	defer lock.Unlock()

	for k := range cache[collection] {
		if strings.HasPrefix(k, prefix) {
			delete(cache[collection], k)
		}
	}

	return nil
}

func ClearAll() error {
	for k := range cache {
		delete(cache, k)
	}

	return nil
}

func Prune() {
	now := time.Now()

	for k := range cache {
		for l := range cache[k] {
			if cache[k][l].Expires.Before(now) {
				delete(cache[k], l)
			}
		}
	}
}

func hashKey(key string) (string, error) {
	return key, nil
}
