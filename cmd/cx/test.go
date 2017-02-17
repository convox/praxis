package main

import (
	"fmt"
	"os"
	"time"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
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
	name := fmt.Sprintf("test-%d", time.Now().Unix())

	app, err := Rack.AppCreate(name)
	if err != nil {
		return err
	}

	defer Rack.AppDelete(name)

	build, err := buildDirectory(app.Name, ".")
	if err != nil {
		return err
	}

	m, err := manifest.LoadFile("convox.yml")
	if err != nil {
		return err
	}

	if err := buildLogs(build, types.Stream{Writer: m.PrefixWriter(os.Stdout, "build")}); err != nil {
		return err
	}

	build, err = Rack.BuildGet(app.Name, build.Id)
	if err != nil {
		return err
	}

	for _, s := range m.Services {
		if s.Test == "" {
			continue
		}

		w := m.PrefixWriter(os.Stdout, s.Name)

		if err := w.Writef("running: %s\n", s.Test); err != nil {
			return err
		}

		err := Rack.ProcessRun(app.Name, types.ProcessRunOptions{
			Command: s.Test,
			Service: s.Name,
			Stream: types.Stream{
				Reader: nil,
				Writer: w,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
