package main

import (
	"os"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "start",
		Description: "start the app in development mode",
		Action:      runStart,
	})
}

func runStart(c *cli.Context) error {
	name := "test"

	app, err := Rack.AppGet(name)
	if err != nil {
		return err
	}

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

		w.Writef("starting\n")

		err := Rack.ProcessRun(app.Name, types.ProcessRunOptions{
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

	// proxy

	// changes

	return nil
}
