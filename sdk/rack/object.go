package rack

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/convox/praxis/types"
)

func (c *Client) ObjectFetch(app string, key string) (io.Reader, error) {
	return c.GetStream(fmt.Sprintf("/apps/%s/objects/%s", app, key))
}

func (c *Client) ObjectStore(app string, key string, r io.Reader) (*types.Object, error) {
	r, err := c.PostStream(fmt.Sprintf("/apps/%s/objects/%s", app, key), r)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var o types.Object

	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}

	return &o, nil
}
