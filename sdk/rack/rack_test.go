package rack_test

import (
	"net/http/httptest"

	"github.com/convox/praxis/cycle"
	"github.com/convox/praxis/sdk/rack"
)

var ts *httptest.Server

func testRack() (rack.Rack, *cycle.HTTP) {
	c, err := cycle.NewHTTP()
	if err != nil {
		panic(err)
	}

	r, err := rack.New(c.Listen())
	if err != nil {
		panic(err)
	}

	return r, c
}

func cleanup() {
	ts.Close()
}

type Server struct {
}
