package rack

import (
	"fmt"
)

func (c *Client) TableFetch(app, table, id string) (attrs map[string]string, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables/%s/id/%s", app, table, id), RequestOptions{}, &attrs)
	return
}

func (c *Client) TableStore(app, table string, attrs map[string]string) (id string, err error) {
	ro := RequestOptions{
		Params: attrs,
	}

	err = c.Post(fmt.Sprintf("/apps/%s/tables/%s", app, table), ro, &id)
	return
}
