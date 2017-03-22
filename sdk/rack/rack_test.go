package rack_test

import (
	"crypto/tls"
	"net/http/httptest"
	"os"

	"github.com/convox/praxis/provider"
	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/server"
	"github.com/convox/praxis/server/controllers"
)

var ts *httptest.Server

const tmpDir = "/tmp/convox"

func setup() (rack.Rack, error) {
	os.Setenv("PROVIDER_ROOT", tmpDir)
	p, err := provider.FromEnv()
	if err != nil {
		return nil, err
	}

	controllers.Provider = p

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
	os.RemoveAll(tmpDir)
}
