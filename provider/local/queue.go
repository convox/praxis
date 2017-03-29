package local

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/convox/praxis/types"
)

func (p *Provider) QueueFetch(app, queue string, opts types.QueueFetchOptions) (map[string]string, error) {
	tick := time.Tick(200 * time.Millisecond)
	timeout := time.After(time.Duration(coalescei(opts.Timeout, 10)) * time.Second)

	for {
		select {
		case <-tick:
			item, err := p.queueFetchItem(app, queue)
			if err != nil {
				return nil, err
			}
			if item != nil {
				return item, nil
			}
		case <-timeout:
			return nil, nil
		}
	}
}

func (p *Provider) QueueStore(app, queue string, attrs map[string]string) error {
	data, err := json.Marshal(attrs)
	if err != nil {
		return err
	}

	return p.storageLogWrite(fmt.Sprintf("apps/%s/queues/%s", app, queue), data)
}

func (p *Provider) queueFetchItem(app, queue string) (map[string]string, error) {
	var item map[string]string

	err := p.storageBucket(fmt.Sprintf("apps/%s/queues/%s", app, queue), func(bucket *bolt.Bucket) error {
		k, v := bucket.Cursor().First()

		if k == nil {
			return nil
		}

		if err := json.Unmarshal(v, &item); err != nil {
			return err
		}

		return bucket.Delete(k)
	})
	if err != nil {
		return nil, err
	}

	return item, nil
}
