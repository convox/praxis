package controllers_test

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

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
	assert.Equal(t, []byte(`{"Name":"test","Release":"","Status":""}`), data)
}

func TestAppList(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	apps := types.Apps{
		{Name: "foo"},
		{Name: "bar"},
	}

	mp.On("AppList").Return(apps, nil)

	res, err := testRequest(ts, "GET", "/apps", nil)
	assert.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, `[{"Name":"foo","Release":"","Status":""},{"Name":"bar","Release":"","Status":""}]`, string(data))
}
