package local

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

func (p *Provider) QueueFetch(app, queue string, opts types.QueueFetchOptions) (map[string]string, error) {
	log := p.logger("QueueFetch").Append("app=%q queue=%q", app, queue)

	tick := time.Tick(200 * time.Millisecond)
	timeout := time.After(time.Duration(coalescei(opts.Timeout, 10)) * time.Second)

	if _, err := p.AppGet(app); err != nil {
		return nil, log.Error(err)
	}

	for {
		select {
		case <-tick:
			item, err := p.queueFetchItem(app, queue)
			if err != nil {
				return nil, errors.WithStack(log.Error(err))
			}
			if item != nil {
				return item, log.Success()
			}
		case <-timeout:
			return nil, log.Successf("timeout=true")
		}
	}
}

func (p *Provider) QueueStore(app, queue string, attrs map[string]string) error {
	log := p.logger("QueueStore").Append("app=%q queue=%q", app, queue)

	if _, err := p.AppGet(app); err != nil {
		return log.Error(err)
	}

	data, err := json.Marshal(attrs)
	if err != nil {
		return errors.WithStack(log.Error(err))
	}

	if err := p.storageLogWrite(fmt.Sprintf("apps/%s/queues/%s", app, queue), data); err != nil {
		return errors.WithStack(log.Error(err))
	}

	return log.Success()
}

func (p *Provider) queueFetchItem(app, queue string) (map[string]string, error) {
	if _, err := p.AppGet(app); err != nil {
		return nil, err
	}

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
