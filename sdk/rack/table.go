package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) TableCreate(app, table string, opts types.TableCreateOptions) error {
	ro := RequestOptions{
		Params: Params{
			"index": opts.Indexes,
		},
	}

	return c.Post(fmt.Sprintf("/apps/%s/tables/%s", app, table), ro, nil)
}

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

func (c *Client) TableGet(app, table string) (m *types.Table, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables/%s", app, table), RequestOptions{}, &m)
	return
}

func (c *Client) TableList(app string) (tables types.Tables, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables", app), RequestOptions{}, &tables)
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

	err = c.Post(fmt.Sprintf("/apps/%s/tables/%s/rows", app, table), ro, &id)
	return
}

func (c *Client) TableRemove(app, table, key string, opts types.TableRemoveOptions) error {
	return c.Delete(fmt.Sprintf("/apps/%s/tables/%s/indexes/%s/%s", app, table, coalesce(opts.Index, "id"), key), RequestOptions{}, nil)
}

func (c *Client) TableRemoveBatch(app, table string, keys []string, opts types.TableRemoveOptions) error {
	ro := RequestOptions{
		Params: Params{
			"key": keys,
		},
	}

	return c.Post(fmt.Sprintf("/apps/%s/tables/%s/indexes/%s/batch/remove", app, table, coalesce(opts.Index, "id")), ro, nil)
}

func (c *Client) TableTruncate(app, table string) error {
	return c.Post(fmt.Sprintf("/apps/%s/tables/%s/truncate", app, table), RequestOptions{}, nil)
}
