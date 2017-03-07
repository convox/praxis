package rack

import (
	"net/url"
	"os"

	"github.com/convox/praxis/provider"
)

type Mock provider.MockProvider

type Rack provider.Provider

func New(host string) Rack {
	return &Client{
		Host: host,
	}
}

func NewFromEnv() (Rack, error) {
	u, err := url.Parse(os.Getenv("RACK_URL"))
	if err != nil {
		return nil, err
	}

	return &Client{
		Host: u.Host,
	}, nil
}
