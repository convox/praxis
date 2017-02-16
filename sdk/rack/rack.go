package rack

import (
	"net/url"
	"os"
)

func New(host string) *Client {
	return &Client{
		Host: host,
	}
}

func NewFromEnv() (*Client, error) {
	u, err := url.Parse(os.Getenv("RACK_URL"))
	if err != nil {
		return nil, err
	}

	return &Client{
		Host: u.Host,
	}, nil
}
