package rack

import (
	"fmt"
	"net/url"

	"github.com/convox/praxis/manifest"
)

func (c *Client) TableFetch(app, table, id string) (attrs map[string]string, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables/%s/id?id=%s", app, table, id), RequestOptions{}, &attrs)
	return
}

func (c *Client) TableFetchIndex(app, table, index, key string) ([]map[string]string, error) {
	return c.TableFetchIndexBatch(app, table, index, []string{key})
}

func (c *Client) TableFetchIndexBatch(app, table, index string, keys []string) (attrs []map[string]string, err error) {
	form := url.Values{}
	form["id"] = keys

	err = c.Get(fmt.Sprintf("/apps/%s/tables/%s/%s?", app, table, index), RequestOptions{UrlForm: form}, &attrs)
	return
}

func (c *Client) TableGet(app, table string) (m *manifest.Table, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/tables/%s", app, table), RequestOptions{}, &m)
	return
}

func (c *Client) TableStore(app, table string, attrs map[string]string) (id string, err error) {
	ro := RequestOptions{
		Params: attrs,
	}

	err = c.Post(fmt.Sprintf("/apps/%s/tables/%s", app, table), ro, &id)
	return
}
