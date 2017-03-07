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
	if err := startLocalRack(); err != nil {
		return err
	}

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

	env, err := manifest.LoadEnvironment(".env")
	if err != nil {
		return err
	}

	if err := m.Validate(env); err != nil {
		return err
	}

	if err := buildLogs(build, types.Stream{Writer: m.Writer("build", os.Stdout)}); err != nil {
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

		w := m.Writer(s.Name, os.Stdout)

		if err := w.Writef("running: %s\n", s.Test); err != nil {
			return err
		}

		senv, err := s.Env(env)
		if err != nil {
			return err
		}

		code, err := Rack.ProcessRun(app.Name, types.ProcessRunOptions{
			Command:     s.Test,
			Environment: senv,
			Service:     s.Name,
			Stream: types.Stream{
				Reader: nil,
				Writer: w,
			},
		})
		if err != nil {
			return err
		}
		if code > 0 {
			return cli.NewExitError(fmt.Sprintf("exit %d", code), code)
		}
	}

	return nil
}
