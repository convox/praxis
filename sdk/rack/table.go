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

func (c *Client) TableGet(app, table string) (m *types.Table, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables/%s", app, table), RequestOptions{}, &m)
	return
}

func (c *Client) TableList(app string) (tables types.Tables, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables", app), RequestOptions{}, &tables)
	return
}

func (c *Client) TableTruncate(app, table string) error {
	return c.Post(fmt.Sprintf("/apps/%s/tables/%s/truncate", app, table), RequestOptions{}, nil)
}

func (c *Client) TableRowDelete(app, table, key string, opts types.TableRowDeleteOptions) error {
	return c.Delete(fmt.Sprintf("/apps/%s/tables/%s/indexes/%s/%s", app, table, coalesce(opts.Index, "id"), key), RequestOptions{}, nil)
}

func (c *Client) TableRowGet(app, table, key string, opts types.TableRowGetOptions) (row *types.TableRow, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables/%s/indexes/%s/%s", app, table, coalesce(opts.Index, "id"), key), RequestOptions{}, &row)
	return
}

func (c *Client) TableRowStore(app, table string, attrs types.TableRow) (id string, err error) {
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

func (c *Client) TableRowsDelete(app, table string, keys []string, opts types.TableRowDeleteOptions) error {
	ro := RequestOptions{
		Params: Params{
			"key": keys,
		},
	}

	return c.Post(fmt.Sprintf("/apps/%s/tables/%s/indexes/%s/batch/remove", app, table, coalesce(opts.Index, "id")), ro, nil)
}

func (c *Client) TableRowsGet(app, table string, keys []string, opts types.TableRowGetOptions) (rows types.TableRows, err error) {
	ro := RequestOptions{
		Params: Params{
			"key": keys,
		},
	}

	err = c.Post(fmt.Sprintf("/apps/%s/tables/%s/indexes/%s/batch", app, table, coalesce(opts.Index, "id")), ro, &rows)
	return
}
