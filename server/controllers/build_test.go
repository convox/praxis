package controllers_test

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildCreate(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	build := &types.Build{
		Id:     "BTEST",
		App:    "app",
		Status: "created",
	}
	opts := types.BuildCreateOptions{Cache: true}
	mp.On("BuildCreate", "app", "http://example.com", opts).Return(build, nil)

	v := url.Values{}
	v.Add("url", "http://example.com")
	v.Add("cache", "true")

	res, err := testRequest(ts, "POST", "/apps/app/builds", bytes.NewReader([]byte(v.Encode())))
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t,
		"{\n  \"id\": \"BTEST\",\n  \"app\": \"app\",\n  \"manifest\": \"\",\n  \"process\": \"\",\n  \"release\": \"\",\n  \"status\": \"created\",\n  \"created\": \"0001-01-01T00:00:00Z\",\n  \"started\": \"0001-01-01T00:00:00Z\",\n  \"ended\": \"0001-01-01T00:00:00Z\"\n}",
		string(data),
	)
}

func TestBuildGet(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	app := &types.App{
		Name: "app",
	}

	mp.On("AppGet", "app").Return(app, nil)

	build := &types.Build{
		Id:     "BTEST",
		App:    "app",
		Status: "created",
	}

	mp.On("BuildGet", "app", "BTEST").Return(build, nil)

	res, err := testRequest(ts, "GET", "/apps/app/builds/BTEST", nil)
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if assert.NoError(t, err) {
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t,
			"{\n  \"id\": \"BTEST\",\n  \"app\": \"app\",\n  \"manifest\": \"\",\n  \"process\": \"\",\n  \"release\": \"\",\n  \"status\": \"created\",\n  \"created\": \"0001-01-01T00:00:00Z\",\n  \"started\": \"0001-01-01T00:00:00Z\",\n  \"ended\": \"0001-01-01T00:00:00Z\"\n}",
			string(data),
		)
	}
}

func TestBuildUpdate(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	app := &types.App{
		Name: "app",
	}

	mp.On("AppGet", "app").Return(app, nil)

	build := &types.Build{
		Id:      "BTEST",
		App:     "app",
		Status:  "pending",
		Release: "RTEST",
	}

	opts := types.BuildUpdateOptions{
		Manifest: `{"manifest": "foo"}`,
		Release:  "RTEST",
		Status:   "pending",
	}

	mp.On("BuildUpdate", "app", "BTEST", opts).Return(build, nil)

	v := url.Values{}
	v.Add("manifest", `{"manifest": "foo"}`)
	v.Add("release", "RTEST")
	v.Add("status", "pending")

	res, err := testRequest(ts, "PUT", "/apps/app/builds/BTEST", bytes.NewReader([]byte(v.Encode())))
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t,
		"{\n  \"id\": \"BTEST\",\n  \"app\": \"app\",\n  \"manifest\": \"\",\n  \"process\": \"\",\n  \"release\": \"RTEST\",\n  \"status\": \"pending\",\n  \"created\": \"0001-01-01T00:00:00Z\",\n  \"started\": \"0001-01-01T00:00:00Z\",\n  \"ended\": \"0001-01-01T00:00:00Z\"\n}",
		string(data),
	)
}
