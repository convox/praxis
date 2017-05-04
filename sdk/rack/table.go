package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (c *Client) TableGet(app, table string) (m *types.Table, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables/%s", app, table), RequestOptions{}, &m)
	return
}

func (c *Client) TableList(app string) (tables types.Tables, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables", app), RequestOptions{}, &tables)
	return
}

func (c *Client) TableQuery(app, table, query string) (rows types.TableRows, err error) {
	return nil, fmt.Errorf("unimplemented")
}

func (c *Client) TableTruncate(app, table string) error {
	return c.Post(fmt.Sprintf("/apps/%s/tables/%s/truncate", app, table), RequestOptions{}, nil)
}
