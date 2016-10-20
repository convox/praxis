package source

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

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
	case "object":
		fmt.Printf("u = %#v\n", u)
		return nil, fmt.Errorf("unsupported")
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
