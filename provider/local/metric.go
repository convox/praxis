package local

import "github.com/convox/praxis/types"

func (p *Provider) MetricList(app, namespace string, opts types.MetricListOptions) ([]string, error) {
	return []string{}, nil
}
func (p *Provider) MetricGet(app, namespace, metric string, opts types.MetricGetOptions) ([]string, error) {
	return []string{}, nil
}
