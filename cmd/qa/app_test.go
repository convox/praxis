// +build qa

package main

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestAppCreate(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppCreate("")
	assert.EqualError(t, err, "app name required")
	assert.Nil(t, app)

	app, err = Rack.AppCreate("3")
	assert.EqualError(t, err, "app name invalid")
	assert.Nil(t, app)

	app, err = appCreate(Rack, "valid")
	assert.NoError(t, err)
	defer appDelete(Rack, "valid")

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

	err = Rack.AppDelete("3")
	assert.EqualError(t, err, "app name invalid")

	err = Rack.AppDelete("missing")
	assert.EqualError(t, err, "no such app: missing")

	app, err := appCreate(Rack, "valid")
	assert.NoError(t, err)
	err = appDelete(Rack, app.Name)
	assert.EqualError(t, err, "no such app: valid")
}

func TestAppGet(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppGet("")
	assert.EqualError(t, err, "response status 404")
	assert.Nil(t, app)

	app, err = Rack.AppGet("3")
	assert.EqualError(t, err, "app name invalid")
	assert.Nil(t, app)

	app, err = appCreate(Rack, "valid")
	assert.NoError(t, err)
	defer appDelete(Rack, "valid")

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

	_, err = appCreate(Rack, "foo")
	assert.NoError(t, err)
	defer appDelete(Rack, "foo")

	_, err = appCreate(Rack, "bar")
	assert.NoError(t, err)
	defer appDelete(Rack, "bar")

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

	app, err := appCreate(Rack, "valid")
	assert.NoError(t, err)
	defer appDelete(Rack, "valid")

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

	app, err := appCreate(Rack, "valid")
	defer appDelete(Rack, app.Name)

	r, err := Rack.AppRegistry(app.Name)
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Hostname)
}

func appCreate(r rack.Rack, name string) (*types.App, error) {
	a, err := r.AppCreate(name)
	if err != nil {
		return a, err
	}

	if err := tickWithTimeout(2*time.Second, 2*time.Minute, notAppStatus(r, name, "creating")); err != nil {
		return nil, err
	}

	return r.AppGet(name)
}

func appDelete(r rack.Rack, name string) error {
	err := r.AppDelete(name)
	if err != nil {
		return err
	}

	if err := tickWithTimeout(2*time.Second, 2*time.Minute, notAppStatus(r, name, "running")); err != nil {
		return err
	}

	if err := tickWithTimeout(2*time.Second, 2*time.Minute, notAppStatus(r, name, "deleting")); err != nil {
		return err
	}

	return nil
}

func notAppStatus(r rack.Rack, app, status string) func() (bool, error) {
	return func() (bool, error) {
		app, err := r.AppGet(app)
		if err != nil {
			return true, err
		}

		if app.Status != status {
			return true, nil
		}

		return false, nil
	}
}

func tickWithTimeout(tick time.Duration, timeout time.Duration, fn func() (stop bool, err error)) error {
	tickch := time.Tick(tick)
	timeoutch := time.After(timeout)

	for {
		stop, err := fn()
		if err != nil {
			return err
		}
		if stop {
			return nil
		}

		select {
		case <-tickch:
			continue
		case <-timeoutch:
			return fmt.Errorf("timeout")
		}
	}
}
