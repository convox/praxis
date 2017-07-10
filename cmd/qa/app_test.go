package main

import (
	"io/ioutil"
	"testing"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestAppCreate(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppCreate("")
	assert.EqualError(t, err, "bucket name required")
	// assert.EqualError(t, err, "app name required") // FIXME
	assert.Nil(t, app)

	app, err = Rack.AppCreate("3")
	defer Rack.AppDelete("3")
	assert.NoError(t, err)
	// assert.EqualError(t, err, "app name invalid") // FIXME
	assert.EqualValues(t, &types.App{
		Name:    "3",
		Release: "",
		Status:  "running",
	}, app)
	// assert.Nil(t, app) // FIXME

	app, err = Rack.AppCreate("valid")
	defer Rack.AppDelete("valid")
	assert.NoError(t, err)
	assert.EqualValues(t, &types.App{
		Name:    "valid",
		Release: "",
		Status:  "running",
	}, app)

	app, err = Rack.AppCreate("valid")
	assert.EqualError(t, err, "app already exists: valid")
	assert.Nil(t, app)
}

func TestAppDelete(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	err = Rack.AppDelete("")
	assert.EqualError(t, err, "response status 404")
	// assert.EqualError(t, err, "app name required") // FIXME

	err = Rack.AppDelete("3")
	assert.EqualError(t, err, "no such app: 3")
	// assert.EqualError(t, err, "app name invalid) // FIXME

	err = Rack.AppDelete("missing")
	assert.EqualError(t, err, "no such app: missing")

	_, err = Rack.AppCreate("valid")
	assert.NoError(t, err)
	err = Rack.AppDelete("valid")
	assert.NoError(t, err)
}

func TestAppGet(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppGet("")
	assert.EqualError(t, err, "response status 404")
	// assert.EqualError(t, err, "app does not exists") // FIXME
	assert.Nil(t, app)

	app, err = Rack.AppCreate("valid")
	defer Rack.AppDelete("valid")
	assert.NoError(t, err)
	a, err := Rack.AppGet("valid")
	assert.NoError(t, err)
	assert.EqualValues(t, app, a)
}

func TestAppList(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	apps, err := Rack.AppList()
	assert.NoError(t, err)
	assert.EqualValues(t, types.Apps{}, apps)

	_, err = Rack.AppCreate("foo")
	defer Rack.AppDelete("foo")
	assert.NoError(t, err)

	_, err = Rack.AppCreate("bar")
	defer Rack.AppDelete("bar")
	assert.NoError(t, err)

	apps, err = Rack.AppList()
	assert.NoError(t, err)
	assert.EqualValues(t, types.Apps{
		types.App{
			Name:    "bar",
			Release: "",
			Status:  "running",
		},
		types.App{
			Name:    "foo",
			Release: "",
			Status:  "running",
		},
	}, apps)
}

func TestAppLogs(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppCreate("valid")
	defer Rack.AppDelete("valid")

	r, err := Rack.AppLogs(app.Name, types.LogsOptions{})
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, b)

	// FIXME: assert app process logs
	// FIXME: assert log options filter, follow, prefix, since
}

func TestAppRegistry(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppCreate("valid")
	defer Rack.AppDelete("valid")

	r, err := Rack.AppRegistry(app.Name)
	assert.NoError(t, err)
	assert.EqualValues(t, &types.Registry{
		Hostname: "convox",
		Password: "",
		Username: "",
	}, r)
}
