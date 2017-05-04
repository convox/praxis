package rack

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/convox/praxis/types"
)

func (c *Client) ObjectExists(app, key string) (exists bool, err error) {
	err = c.Head(fmt.Sprintf("/apps/%s/objects/%s", app, key), RequestOptions{}, &exists)
	return
}

func (c *Client) ObjectFetch(app string, key string) (io.ReadCloser, error) {
	res, err := c.GetStream(fmt.Sprintf("/apps/%s/objects/%s", app, key), RequestOptions{})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) ObjectStore(app string, key string, r io.Reader, opts types.ObjectStoreOptions) (*types.Object, error) {
	ro := RequestOptions{
		Body: r,
	}

	res, err := c.PostStream(fmt.Sprintf("/apps/%s/objects/%s", app, key), ro)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var o types.Object

	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}

	return &o, nil
}
