package aws

import (
	"github.com/convox/praxis/types"
)

func (p *Provider) QueueFetch(app, queue string, opts types.QueueFetchOptions) (map[string]string, error) {
	return nil, nil
}

func (p *Provider) QueueStore(app, queue string, attrs map[string]string) error {
	return nil
}
