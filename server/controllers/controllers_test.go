package controllers_test

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/convox/praxis/server"
	"github.com/convox/praxis/server/controllers"
	"github.com/convox/praxis/types"
)

func mockServer() (*httptest.Server, *types.MockProvider) {
	mp := &types.MockProvider{}
	controllers.Provider = mp

	s := server.New()

	ts := httptest.NewUnstartedServer(s)

	ts.TLS = &tls.Config{
		NextProtos: []string{"h2"},
	}

	ts.StartTLS()

	return ts, mp
}

func testRequest(ts *httptest.Server, method, path string, r io.Reader) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest(method, ts.URL+path, r)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return client.Do(req)
}
