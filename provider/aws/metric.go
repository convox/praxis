package aws

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) MetricGet(app, namespace, metric string, opts types.MetricGetOptions) ([]string, error) {
	return []string{}, nil
}

func (p *Provider) MetricList(app, namespace string) ([]string, error) {
	if metrics, ok := types.MetricNames[namespace]; ok {
		return metrics, nil
	}

	return []string{}, fmt.Errorf("Namespace %s not found", namespace)
}
