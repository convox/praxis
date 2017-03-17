package rack

import (
	"net/url"
	"os"

	"github.com/convox/praxis/provider"
)

const (
	sortableTime = "20060102.150405.000000000"
)

type Mock struct {
	provider.MockProvider
}

type Rack provider.Provider

func New(host string) Rack {
	return &Client{
		Host: host,
	}
}

func NewFromEnv() (Rack, error) {
	u, err := url.Parse(coalesce(os.Getenv("RACK_URL"), "https://localhost:5443"))
	if err != nil {
		return nil, err
	}

	return &Client{
		Host: u.Host,
	}, nil
}
