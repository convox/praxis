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

	name := fmt.Sprintf("test-%d", time.Now().Unix())
	app, err = appCreate(Rack, name)
	assert.NoError(t, err)
	defer appDelete(Rack, name)

	assert.EqualValues(t, &types.App{
		Name:    name,
		Release: "",
		Status:  "running",
	}, app)

	app, err = Rack.AppCreate(name)
	assert.EqualError(t, err, fmt.Sprintf("app already exists: %s", name))
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

	name := fmt.Sprintf("test-%d", time.Now().Unix())
	_, err = appCreate(Rack, name)
	assert.NoError(t, err)
	err = appDelete(Rack, name)
	assert.EqualError(t, err, fmt.Sprintf("no such app: %s", name))
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

	name := fmt.Sprintf("test-%d", time.Now().Unix())
	app, err = appCreate(Rack, name)
	assert.NoError(t, err)
	defer appDelete(Rack, name)

	a, err := Rack.AppGet(name)
	assert.NoError(t, err)
	assert.EqualValues(t, app, a)
}

func TestAppList(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	apps, err := Rack.AppList()
	assert.NoError(t, err)
	assert.EqualValues(t, types.Apps{}, apps)

	name1 := fmt.Sprintf("foo-%d", time.Now().Unix())
	_, err = appCreate(Rack, name1)
	assert.NoError(t, err)
	defer appDelete(Rack, name1)

	name2 := fmt.Sprintf("bar-%d", time.Now().Unix())
	_, err = appCreate(Rack, name2)
	assert.NoError(t, err)
	defer appDelete(Rack, name2)

	apps, err = Rack.AppList()
	assert.NoError(t, err)
	assert.EqualValues(t, types.Apps{
		types.App{
			Name:    name2,
			Release: "",
			Status:  "running",
		},
		types.App{
			Name:    name1,
			Release: "",
			Status:  "running",
		},
	}, apps)
}
func TestAppLogs(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	name := fmt.Sprintf("test-%d", time.Now().Unix())
	_, err = appCreate(Rack, name)
	assert.NoError(t, err)
	defer appDelete(Rack, name)

	r, err := Rack.AppLogs(name, types.LogsOptions{})
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

	name := fmt.Sprintf("test-%d", time.Now().Unix())
	_, err = appCreate(Rack, name)
	defer appDelete(Rack, name)

	r, err := Rack.AppRegistry(name)
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
