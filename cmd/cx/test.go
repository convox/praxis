package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/server"
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "test",
		Description: "run tests",
		Action:      runTest,
	})
}

func runTest(c *cli.Context) error {
	os.Remove("/tmp/test.sock")
	go server.New().Listen("unix", "/tmp/test.sock")
	time.Sleep(100 * time.Millisecond)
	Rack.Socket = "/tmp/test.sock"

	app, err := Rack.AppCreate("test")
	if err != nil {
		return err
	}

	fmt.Printf("app = %+v\n", app)

	release, err := releaseDirectory(app.Name, ".")
	if err != nil {
		return err
	}

	fmt.Printf("release = %+v\n", release)

	data, err := Rack.ReleaseManifest(app.Name, release.Id)
	if err != nil {
		return err
	}

	fmt.Printf("len(data) = %+v\n", len(data))

	m, err := manifest.Load(data)
	if err != nil {
		return err
	}

	fmt.Printf("m = %+v\n", m)

	for _, s := range m.Services {
		if s.Test != "" {
			_, err := Rack.ProcessRun(app.Name, s.Name, rack.ProcessRunOptions{
				Output: os.Stdout,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func releaseDirectory(app, dir string) (*rack.Release, error) {
	context := bytes.NewReader([]byte{})

	object, err := Rack.ObjectStore("", context)
	if err != nil {
		return nil, err
	}

	build, err := Rack.BuildCreate(app, object.URL)
	if err != nil {
		return nil, err
	}

	return Rack.ReleaseCreate(app, rack.ReleaseCreateOptions{
		Build: build.Id,
	})
}
