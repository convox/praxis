package main

import (
	"bytes"
	"fmt"
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
	assert.EqualError(t, err, "Image or service is required")
	assert.Equal(t, 255, code)

	code, err = Rack.ProcessRun(app.Name, types.ProcessRunOptions{
		Output:  logs,
		Service: "web",
	})
	assert.EqualError(t, err, "[no release promoted for app: valid]")
	assert.Equal(t, 255, code)

	bs, err := ioutil.ReadAll(logs)
	assert.NoError(t, err)
	out := string(bs)

	assert.Equal(t, "", out)

	r, err := Rack.ReleaseCreate(app.Name, types.ReleaseCreateOptions{
		Env: map[string]string{
			"FOO": "bar",
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Id)

	logs = bytes.NewBuffer([]byte{})
	code, err = Rack.ProcessRun(app.Name, types.ProcessRunOptions{
		Output:  logs,
		Service: "web",
	})
	assert.EqualError(t, err, "[no release promoted for app: valid]")
	// assert.EqualError(t, err, "[no builds for app: valid]") // FIXME
	assert.Equal(t, 255, code)

	obj, err := Rack.ObjectStore(app.Name, "", tarReader(), types.ObjectStoreOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, obj.Key)

	b, err := Rack.BuildCreate(app.Name, fmt.Sprintf("object:///%s", obj.Key), types.BuildCreateOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, b.Id)

	l, err := Rack.BuildLogs(app.Name, b.Id)
	assert.NoError(t, err)
	_, err = ioutil.ReadAll(l)
	assert.NoError(t, err)

	b, err = Rack.BuildGet(app.Name, b.Id)
	assert.NoError(t, err)

	err = Rack.ReleasePromote(app.Name, b.Release)
	assert.NoError(t, err)

	code, err = Rack.ProcessRun(app.Name, types.ProcessRunOptions{
		Output:  logs,
		Service: "thunk",
	})
	assert.EqualError(t, err, "[no such service: thunk]")
	assert.Equal(t, 255, code)

	logs = bytes.NewBuffer([]byte{})
	code, err = Rack.ProcessRun(app.Name, types.ProcessRunOptions{
		Command: "echo hi",
		Output:  logs,
		Service: "web",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	bs, err = ioutil.ReadAll(logs)
	out = string(bs)
	assert.NoError(t, err)
	assert.Equal(t, "hi\r\n", out)
}

func TestProcessRunReleases(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppCreate("valid")
	defer Rack.AppDelete(app.Name)
	assert.NoError(t, err)

	r, err := Rack.ReleaseCreate(app.Name, types.ReleaseCreateOptions{
		Env: map[string]string{
			"FOO": "bar",
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Id)
	r1 := r.Id

	obj, err := Rack.ObjectStore(app.Name, "", tarReader(File{"a.txt", "a"}), types.ObjectStoreOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, obj.Key)

	b, err := Rack.BuildCreate(app.Name, fmt.Sprintf("object:///%s", obj.Key), types.BuildCreateOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, b.Id)

	l, err := Rack.BuildLogs(app.Name, b.Id)
	assert.NoError(t, err)
	_, err = ioutil.ReadAll(l)
	assert.NoError(t, err)

	b, err = Rack.BuildGet(app.Name, b.Id)
	assert.NoError(t, err)
	assert.NotEmpty(t, b.Release)
	r2 := b.Release

	r, err = Rack.ReleaseCreate(app.Name, types.ReleaseCreateOptions{
		Env: map[string]string{
			"FOO": "baz",
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Id)
	r3 := r.Id

	obj, err = Rack.ObjectStore(app.Name, "", tarReader(File{"b.txt", "b"}), types.ObjectStoreOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, obj.Key)

	b, err = Rack.BuildCreate(app.Name, fmt.Sprintf("object:///%s", obj.Key), types.BuildCreateOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, b.Id)

	l, err = Rack.BuildLogs(app.Name, b.Id)
	assert.NoError(t, err)
	_, err = ioutil.ReadAll(l)
	assert.NoError(t, err)

	b, err = Rack.BuildGet(app.Name, b.Id)
	assert.NoError(t, err)
	assert.NotEmpty(t, b.Release)
	r4 := b.Release

	logs := bytes.NewBuffer([]byte{})
	code, err := Rack.ProcessRun(app.Name, types.ProcessRunOptions{
		Command: "env ; ls",
		Output:  logs,
		Service: "web",
	})
	assert.EqualError(t, err, "[no release promoted for app: valid]")
	assert.Equal(t, 255, code)

	logs = bytes.NewBuffer([]byte{})
	code, err = Rack.ProcessRun(app.Name, types.ProcessRunOptions{
		Command: "env ; ls",
		Output:  logs,
		Release: r1,
		Service: "web",
	})
	assert.EqualError(t, err, "[no builds for app: valid]")
	assert.Equal(t, 255, code)

	bs, err := ioutil.ReadAll(logs)
	out := string(bs)
	assert.NoError(t, err)
	assert.Equal(t, "", out)

	logs = bytes.NewBuffer([]byte{})
	code, err = Rack.ProcessRun(app.Name, types.ProcessRunOptions{
		Command: "env ; ls",
		Output:  logs,
		Release: r2,
		Service: "web",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	bs, err = ioutil.ReadAll(logs)
	out = string(bs)
	assert.Contains(t, out, "FOO=bar")
	assert.Contains(t, out, "a.txt")

	logs = bytes.NewBuffer([]byte{})
	code, err = Rack.ProcessRun(app.Name, types.ProcessRunOptions{
		Command: "env ; ls",
		Output:  logs,
		Release: r3,
		Service: "web",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	bs, err = ioutil.ReadAll(logs)
	out = string(bs)
	assert.Contains(t, out, "FOO=baz")
	assert.Contains(t, out, "a.txt")

	logs = bytes.NewBuffer([]byte{})
	code, err = Rack.ProcessRun(app.Name, types.ProcessRunOptions{
		Command: "env ; ls",
		Output:  logs,
		Release: r4,
		Service: "web",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	bs, err = ioutil.ReadAll(logs)
	out = string(bs)
	assert.Contains(t, out, "FOO=baz")
	assert.Contains(t, out, "b.txt")
	assert.NotContains(t, out, "a.txt")
}
