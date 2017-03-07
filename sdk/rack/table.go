package rack

import (
	"fmt"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func (c *Client) TableFetch(app, table, key string, opts types.TableFetchOptions) (attrs map[string]string, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables/%s/indexes/%s/%s", app, table, coalesce(opts.Index, "id"), key), RequestOptions{}, &attrs)
	return
}

func (c *Client) TableFetchBatch(app, table string, keys []string, opts types.TableFetchOptions) (items []map[string]string, err error) {
	ro := RequestOptions{
		Params: Params{
			"key": keys,
		},
	}

	err = c.Post(fmt.Sprintf("/apps/%s/tables/%s/indexes/%s/batch", app, table, coalesce(opts.Index, "id")), ro, &items)
	return
}

func (c *Client) TableGet(app, table string) (m *manifest.Table, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables/%s", app, table), RequestOptions{}, &m)
	return
}

func (c *Client) TableStore(app, table string, attrs map[string]string) (id string, err error) {
	params := map[string]interface{}{}

	for k, v := range attrs {
		params[k] = v
	}

	ro := RequestOptions{
		Params: params,
	}

	err = c.Post(fmt.Sprintf("/apps/%s/tables/%s", app, table), ro, &id)
	return
}
