package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) MetricList(app, namespace string, opts types.MetricListOptions) ([]string, error) {
	metrics := []string{}
	err := c.Get(fmt.Sprintf("/apps/%s/metrics/%s", app, namespace), RequestOptions{}, &metrics)
	return metrics, err
}

func (c *Client) MetricGet(app, namespace, metric string, opts types.MetricGetOptions) ([]string, error) {
	return []string{}, nil
}
