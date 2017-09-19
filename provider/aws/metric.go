package aws

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) MetricGet(app, name string) (*types.Metric, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) MetricList(app string) (types.Metrics, error) {
	return nil, fmt.Errorf("unimplemented")
}
