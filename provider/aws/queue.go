package aws

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) QueueFetch(app, queue string, opts types.QueueFetchOptions) (map[string]string, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) QueueStore(app, queue string, attrs map[string]string) error {
	return fmt.Errorf("unimplemented")
}
