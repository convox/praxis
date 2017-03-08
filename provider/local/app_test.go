package local_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppCreate(t *testing.T) {
	local, err := Provider()
	assert.NoError(t, err)

	app, err := local.AppCreate("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", app.Name)
}

func TestAppDelete(t *testing.T) {
	local, err := Provider()
	assert.NoError(t, err)

	_, err = local.AppCreate("test")
	assert.NoError(t, err)

	_, err = local.AppGet("test")
	assert.NoError(t, err)

	err = local.AppDelete("test")
	assert.NoError(t, err)

	_, err = local.AppGet("test")
	assert.EqualError(t, err, "no such app: test")
}

func TestAppGet(t *testing.T) {
	local, err := Provider()
	assert.NoError(t, err)

	_, err = local.AppGet("test")
	assert.EqualError(t, err, "no such app: test")

	_, err = local.AppCreate("test")
	assert.NoError(t, err)

	app, err := local.AppGet("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", app.Name)
}

func TestAppList(t *testing.T) {
	local, err := Provider()
	assert.NoError(t, err)

	apps, err := local.AppList()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(apps))

	local.AppCreate("test1")
	local.AppCreate("test3")
	local.AppCreate("test2")

	apps, err = local.AppList()
	assert.NoError(t, err)

	if assert.Equal(t, 3, len(apps)) {
		assert.Equal(t, "test1", apps[0].Name)
		assert.Equal(t, "test2", apps[1].Name)
		assert.Equal(t, "test3", apps[2].Name)
	}
}
