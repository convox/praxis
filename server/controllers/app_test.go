package controllers_test

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/convox/praxis/provider"
	"github.com/convox/praxis/server"
	"github.com/convox/praxis/server/controllers"
	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PROVIDER_ROOT", "/tmp/convox/foo")
}

func mockServer() (*httptest.Server, *provider.MockProvider) {
	mp := &provider.MockProvider{}
	controllers.Provider = mp

	s := server.New()

	ts := httptest.NewUnstartedServer(s)

	ts.TLS = &tls.Config{
		NextProtos: []string{"h2"},
	}

	ts.StartTLS()

	return ts, mp
}

func testRequest(ts *httptest.Server, method, path string, r io.Reader) ([]byte, error) {
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

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}

func TestAppCreate(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	app := &types.App{
		Name: "test",
	}

	mp.On("AppCreate", "test").Return(app, nil)

	v := url.Values{}
	v.Add("name", "test")

	data, err := testRequest(ts, "POST", "/apps", bytes.NewReader([]byte(v.Encode())))
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"Name":"test"}`), data)
}
