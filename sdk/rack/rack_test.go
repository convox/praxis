package rack_test

import (
	"crypto/tls"
	"net/http/httptest"
	"os"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/server"
)

var ts *httptest.Server

func setup() (rack.Rack, error) {
	ts = httptest.NewUnstartedServer(server.New())
	ts.TLS = &tls.Config{
		NextProtos: []string{"h2"},
	}

	ts.StartTLS()

	os.Setenv("RACK_URL", ts.URL)
	return rack.NewFromEnv()
}

func cleanup() {
	ts.Close()
}
