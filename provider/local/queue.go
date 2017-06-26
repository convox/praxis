package local

import (
	"fmt"
	"time"

	"github.com/convox/praxis/types"
)

type queueChannel chan map[string]string

var (
	queues = map[string]queueChannel{}
)

func (p *Provider) QueueFetch(app, queue string, opts types.QueueFetchOptions) (map[string]string, error) {
	log := p.logger("QueueFetch").Append("app=%q queue=%q", app, queue)

	timeout := time.After(time.Duration(coalescei(opts.Timeout, 10)) * time.Second)

	if _, err := p.AppGet(app); err != nil {
		return nil, log.Error(err)
	}

	select {
	case attrs := <-appQueue(app, queue):
		return attrs, log.Success()
	case <-timeout:
		return nil, log.Successf("timeout=true")
	}
}

func (p *Provider) QueueStore(app, queue string, attrs map[string]string) error {
	log := p.logger("QueueStore").Append("app=%q queue=%q", app, queue)

	if _, err := p.AppGet(app); err != nil {
		return log.Error(err)
	}

	appQueue(app, queue) <- attrs

	return log.Success()
}

func appQueue(app, queue string) queueChannel {
	key := fmt.Sprintf("%s-%s", app, queue)

	if q, ok := queues[key]; ok {
		return q
	}

	queues[key] = make(queueChannel, 10*1024)

	return queues[key]
}
