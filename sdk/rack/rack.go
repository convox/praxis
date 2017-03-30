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

func New(endpoint string) (Rack, error) {
	u, err := url.Parse(coalesce(endpoint, "https://localhost:5443"))
	if err != nil {
		return nil, err
	}

	return &Client{Endpoint: u}, nil
}

func NewFromEnv() (Rack, error) {
	return New(os.Getenv("RACK_URL"))
}
