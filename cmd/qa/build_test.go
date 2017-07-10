package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildCreate(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppCreate("valid")
	defer Rack.AppDelete("valid")
	assert.NoError(t, err)

	build, err := Rack.BuildCreate("invalid", "object:///thunk", types.BuildCreateOptions{})
	assert.EqualError(t, err, "no such app: invalid")

	build, err = Rack.BuildCreate("valid", "", types.BuildCreateOptions{})
	assert.NoError(t, err)
	// assert.EqualError(t, err, "object is required") // FIXME

	build, err = Rack.BuildCreate("valid", "invalid", types.BuildCreateOptions{})
	assert.NoError(t, err)
	// assert.EqualError(t, err, "object url is invalid") // FIXME

	build, err = Rack.BuildCreate("valid", "object://missing", types.BuildCreateOptions{})
	assert.NoError(t, err)
	// assert.EqualError(t, err, "object does not exist") // FIXME

	obj, err := Rack.ObjectStore(app.Name, "", tarReader(), types.ObjectStoreOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, obj.Key)

	build, err = Rack.BuildCreate(app.Name, fmt.Sprintf("object:///%s", obj.Key), types.BuildCreateOptions{})
	assert.NoError(t, err)
	assert.Regexp(t, "B[A-Z]{9}", build.Id)
	assert.Equal(t, "valid", build.App)
	assert.Equal(t, "", build.Manifest)
	assert.Regexp(t, "[a-z0-9]{64}", build.Process)
	assert.Equal(t, "", build.Release)
	assert.Equal(t, "created", build.Status)
	assert.WithinDuration(t, time.Now(), build.Created, 2*time.Minute)
	// assert.WithinDuration(t, time.Now(), build.Created, 5*time.Second) // FIXME time skew?
	assert.Equal(t, time.Time{}, build.Started)
	assert.Equal(t, time.Time{}, build.Ended)
}

func TestBuildCreateOptions(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppCreate("valid")
	defer Rack.AppDelete("valid")
	assert.NoError(t, err)

	obj, err := Rack.ObjectStore(app.Name, "", tarReader(), types.ObjectStoreOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, obj.Key)

	b, err := Rack.BuildCreate(app.Name, fmt.Sprintf("object:///%s", obj.Key), types.BuildCreateOptions{
		Manifest: "missing.yml",
	})
	assert.NoError(t, err)

	logs, err := Rack.BuildLogs(app.Name, b.Id)
	assert.NoError(t, err)

	bytes, err := ioutil.ReadAll(logs)
	assert.NoError(t, err)
	out := string(bytes)

	assert.Contains(t, out, "pulling: httpd")
	// assert.Contains(t, out, "missing.yml: no such file or directory") // FIXME

	b, err = Rack.BuildCreate(app.Name, fmt.Sprintf("object:///%s", obj.Key), types.BuildCreateOptions{
		Cache: false,
	})
	assert.NoError(t, err)

	logs, err = Rack.BuildLogs(app.Name, b.Id)
	assert.NoError(t, err)

	bytes, err = ioutil.ReadAll(logs)
	assert.NoError(t, err)
	out = string(bytes)

	assert.Contains(t, out, "Status: Image is up to date for httpd:latest")
	// assert.Contains(t, out, "Status: Downloaded newer image for httpd:latest") // FIXME
}

func TestBuildGetLogs(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppCreate("valid")
	defer Rack.AppDelete("valid")
	assert.NoError(t, err)

	obj, err := Rack.ObjectStore(app.Name, "", tarReader(), types.ObjectStoreOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, obj.Key)

	b, err := Rack.BuildCreate(app.Name, fmt.Sprintf("object:///%s", obj.Key), types.BuildCreateOptions{})
	assert.NoError(t, err)

	// FIXME: not guaranteed to be running
	build, err := Rack.BuildGet(app.Name, b.Id)
	assert.Equal(t, b.Id, build.Id)
	assert.Equal(t, b.App, build.App)
	assert.Equal(t, b.Manifest, build.Manifest)
	assert.Regexp(t, b.Process, build.Process)
	assert.Equal(t, b.Release, build.Release)
	assert.Equal(t, "running", build.Status)
	assert.Equal(t, b.Created, build.Created)
	assert.WithinDuration(t, time.Now(), build.Started, 2*time.Minute)
	assert.Equal(t, time.Time{}, build.Ended)

	logs, err := Rack.BuildLogs(app.Name, b.Id)
	assert.NoError(t, err)

	bytes, err := ioutil.ReadAll(logs)
	assert.NoError(t, err)
	out := string(bytes)

	assert.NoError(t, err)
	assert.Contains(t, out, "preparing source")
	assert.Contains(t, out, "pulling: httpd")
	assert.Contains(t, out, "running: docker pull httpd")
	assert.Contains(t, out, fmt.Sprintf("running: docker tag httpd convox/valid/web:%s", b.Id))
	assert.Contains(t, out, "storing artifacts")
	assert.Contains(t, out, "build complete")

	build, err = Rack.BuildGet(app.Name, b.Id)
	assert.Equal(t, b.Id, build.Id)
	assert.Equal(t, b.App, build.App)
	assert.Equal(t, "services:\n  web:\n    image: httpd\n    port: 80", build.Manifest)
	assert.Regexp(t, b.Process, build.Process)
	assert.Regexp(t, "R[A-Z]{9}", build.Release)
	assert.Equal(t, "complete", build.Status)
	assert.Equal(t, b.Created, build.Created)
	assert.WithinDuration(t, time.Now(), build.Started, 2*time.Minute)
	assert.WithinDuration(t, time.Now(), build.Ended, 2*time.Minute)
}

func TestBuildList(t *testing.T) {
	Rack, err := rack.NewFromEnv()
	assert.NoError(t, err)

	app, err := Rack.AppCreate("valid")
	defer Rack.AppDelete("valid")
	assert.NoError(t, err)

	ids := []string{}
	for i := 0; i < 11; i++ {
		b, err := Rack.BuildCreate(app.Name, "", types.BuildCreateOptions{})
		assert.NoError(t, err)
		ids = append(ids, b.Id)
	}

	builds, err := Rack.BuildList(app.Name)
	assert.NoError(t, err)

	assert.Len(t, builds, 11)
	// assert.Len(t, builds, 11) // FIXME should be paginated
	assert.Equal(t, ids[10], builds[0].Id)
	assert.Equal(t, ids[0], builds[10].Id)
}

func tarReader() io.Reader {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	var files = []struct {
		Name, Body string
	}{
		{"convox.yml", `services:
  web:
    image: httpd
    port: 80`},
		{"README.md", "hello"},
	}

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatalln(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatalln(err)
		}
	}

	// Make sure to check the error on Close.
	if err := tw.Close(); err != nil {
		log.Fatalln(err)
	}

	// Open the tar archive for reading.
	return bytes.NewReader(buf.Bytes())
}

func assertStructMatches(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	return assert.EqualValues(t, expected, actual, msgAndArgs...)
}
