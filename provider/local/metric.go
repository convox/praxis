package local

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) MetricList(app, namespace string, opts types.MetricListOptions) ([]string, error) {
	if metrics, ok := types.MetricNames[namespace]; ok {
		return metrics, nil
	}

	return []string{}, fmt.Errorf("Namespace %s not found", namespace)
}
func (p *Provider) MetricGet(app, namespace, metric string, opts types.MetricGetOptions) ([]string, error) {
	return []string{}, nil
}
