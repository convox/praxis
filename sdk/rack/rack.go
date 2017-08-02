package rack

import (
	"net/url"
	"os"

	"github.com/convox/praxis/types"
)

const (
	sortableTime = "20060102.150405.000000000"
)

type Rack types.Provider

func New(endpoint string) (*Client, error) {
	u, err := url.Parse(coalesce(endpoint, "https://localhost:5443"))
	if err != nil {
		return nil, err
	}

	return &Client{Debug: os.Getenv("CONVOX_DEBUG") == "true", Endpoint: u, Version: "dev"}, nil
}

func NewFromEnv() (*Client, error) {
	return New(os.Getenv("RACK_URL"))
}
