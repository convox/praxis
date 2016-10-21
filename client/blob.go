package client

import (
	"fmt"
	"io"
)

type BlobStoreOptions struct {
	Public bool
}

func (c *Client) BlobFetch(app, key string) (io.ReadCloser, error) {
	return c.GetReader(fmt.Sprintf("/apps/%s/blobs/%s", app, key), GetOptions{})
}

func (c *Client) BlobStore(app, key string, r io.Reader, opts BlobStoreOptions) (string, error) {
	var url string

	popts := PostOptions{
		Files: map[string]io.Reader{
			"data": r,
		},
	}

	if err := c.Post(fmt.Sprintf("/apps/%s/blobs/%s", app, key), &url, popts); err != nil {
		return "", err
	}

	return url, nil
}
