package rack

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

func (c *Client) KeyDecrypt(app, key string, data []byte) ([]byte, error) {
	ro := RequestOptions{
		Body: bytes.NewReader(data),
	}

	res, err := c.PostStream(fmt.Sprintf("/apps/%s/keys/%s/decrypt", app, key), ro)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Client) KeyEncrypt(app, key string, data []byte) ([]byte, error) {
	ro := RequestOptions{
		Body: bytes.NewReader(data),
	}

	res, err := c.PostStream(fmt.Sprintf("/apps/%s/keys/%s/encrypt", app, key), ro)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return out, nil
}
