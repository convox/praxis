package source

import (
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/convox/praxis/client"
	"github.com/convox/praxis/provider"
	"github.com/convox/praxis/provider/local"
)

type Source interface {
	Fetch(out io.Writer) (string, error)
}

func urlReader(url_ string) (io.ReadCloser, error) {
	u, err := url.Parse(url_)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "file":
		fd, err := os.Open(u.Path)
		if err != nil {
			return nil, err
		}
		return fd, nil
	case "blob":
		return client.New(os.Getenv("CONVOX_URL")).BlobFetch(os.Getenv("BUILD_APP"), u.Path)
		// return providerFromEnv().BlobFetch(app, u.Path)
	}

	req, err := http.Get(url_)
	if err != nil {
		return nil, err
	}

	return req.Body, nil
}

func providerFromEnv() provider.Provider {
	switch os.Getenv("PROVIDER") {
	default:
		return local.FromEnv()
	}
}
