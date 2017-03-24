package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) MetricGet(app, namespace, metric string, opts types.MetricGetOptions) ([]string, error) {
	metrics := []string{}
	err := c.Get(fmt.Sprintf("/apps/%s/metrics/%s/%s", app, namespace, metric), RequestOptions{}, &metrics)
	return metrics, err
}

func (c *Client) MetricList(app, namespace string) ([]string, error) {
	metrics := []string{}
	err := c.Get(fmt.Sprintf("/apps/%s/metrics/%s", app, namespace), RequestOptions{}, &metrics)
	return metrics, err
}
