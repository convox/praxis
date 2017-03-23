package rack

import "github.com/convox/praxis/types"

func (c *Client) MetricList(app, namespace string, opts types.MetricListOptions) ([]string, error) {
	return []string{}, nil
}
func (c *Client) MetricGet(app, namespace, metric string, opts types.MetricGetOptions) ([]string, error) {
	return []string{}, nil
}
