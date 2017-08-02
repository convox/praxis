package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) MetricGet(app, name string) (metric *types.Metric, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/metrics/%s", app, name), RequestOptions{}, &metric)
	return
}

func (c *Client) MetricList(app string) (metrics types.Metrics, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/metrics", app), RequestOptions{}, &metrics)
	return
}
