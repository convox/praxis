package controllers_test

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PROVIDER_ROOT", "/tmp/convox/foo")
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

	res, err := testRequest(ts, "POST", "/apps", bytes.NewReader([]byte(v.Encode())))
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, []byte(`{"Name":"test"}`), data)
}
