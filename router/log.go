package router

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/convox/praxis/sdk/rack"
)

type logTransport struct {
	*http.Transport
	listener *url.URL
	rack     rack.Rack
}

func (t logTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("ns=convox.router at=proxy type=http listen=%q target=%q\n", t.listener, req.URL)

	if req.URL.Hostname() == "rack" {
	}

	return t.Transport.RoundTrip(req)
}
