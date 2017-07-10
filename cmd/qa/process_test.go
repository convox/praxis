package main

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestProcessRun(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppCreate("valid")
	defer Rack.AppDelete(app.Name)
	assert.NoError(t, err)

	code, err := Rack.ProcessRun(app.Name, types.ProcessRunOptions{})
	assert.EqualError(t, err, "Output is required")
	assert.Equal(t, 255, code)

	logs := bytes.NewBuffer([]byte{})
	code, err = Rack.ProcessRun(app.Name, types.ProcessRunOptions{
		Output: logs,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	bytes, err := ioutil.ReadAll(logs)
	assert.NoError(t, err)
	out := string(bytes)

	assert.Equal(t, "26bda8cd-ad49-4e4b-8bb3-2f19e197b3bd[-1][no releases for app: valid]", out)
}
