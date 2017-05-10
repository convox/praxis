package main

import (
	"fmt"
	"os"
	"strings"
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
		Action:      errorExit(runTest, 1),
	})
}

func runTest(c *cli.Context) error {
	name := fmt.Sprintf("test-%d", time.Now().Unix())

	app, err := Rack.AppCreate(name)
	if err != nil {
		return err
	}

	defer Rack.AppDelete(name)

	if err := tickWithTimeout(2*time.Second, 1*time.Minute, notAppStatus(name, "creating")); err != nil {
		return err
	}

	m, err := manifest.LoadFile("convox.yml")
	if err != nil {
		return err
	}

	env := map[string]string{}

	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)

		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}

	if err := m.Validate(env); err != nil {
		return err
	}

	bw := types.Stream{Writer: m.Writer("build", os.Stdout)}

	build, err := buildDirectory(app.Name, ".", bw)
	if err != nil {
		return err
	}

	if err := buildLogs(build, bw); err != nil {
		return err
	}

	for _, s := range m.Services {
		if s.Command.Test == "" {
			continue
		}

		w := m.Writer(s.Name, os.Stdout)

		if err := w.Writef("running: %s\n", s.Command.Test); err != nil {
			return err
		}

		senv, err := s.Env(env)
		if err != nil {
			return err
		}

		code, err := Rack.ProcessRun(app.Name, types.ProcessRunOptions{
			Command:     s.Command.Test,
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
