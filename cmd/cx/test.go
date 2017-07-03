package main

import (
	"fmt"
	"io/ioutil"
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
		Before:      beforeCmd,
	})
}

func runTest(c *cli.Context) error {
	env := manifest.Environment{}

	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)

		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}

	data, err := ioutil.ReadFile("convox.yml")
	if err != nil {
		return err
	}

	m, err := manifest.Load(data, env)
	if err != nil {
		return err
	}

	system := m.Writer("convox", os.Stdout)

	stdcli.DefaultWriter.Stdout = system
	stdcli.DefaultWriter.Stderr = system

	name := fmt.Sprintf("test-%d", time.Now().Unix())

	stdcli.Startf("creating app <name>%s</name>", name)

	app, err := Rack(c).AppCreate(name)
	if err != nil {
		return err
	}

	defer Rack(c).AppDelete(name)

	if err := tickWithTimeout(2*time.Second, 1*time.Minute, notAppStatus(Rack(c), name, "creating")); err != nil {
		return err
	}

	stdcli.OK()

	if err := m.Validate(); err != nil {
		return err
	}

	build, err := buildDirectory(Rack(c), app.Name, ".", types.BuildCreateOptions{Development: true}, m.Writer("build", os.Stdout))
	if err != nil {
		return err
	}

	if err := Rack(c).ReleasePromote(app.Name, build.Release); err != nil {
		return err
	}

	if err := releaseLogs(Rack(c), app.Name, build.Release, m.Writer("release", os.Stdout), types.LogsOptions{Follow: true}); err != nil {
		return err
	}

	r, err := Rack(c).ReleaseGet(app.Name, build.Release)
	if err != nil {
		return err
	}

	if r.Status != "promoted" {
		return fmt.Errorf("promote failed")
	}

	for _, s := range m.Services {
		if s.Test == "" {
			continue
		}

		w := m.Writer(s.Name, os.Stdout)

		if err := w.Writef("running: %s\n", s.Test); err != nil {
			return err
		}

		senv, err := m.ServiceEnvironment(s.Name)
		if err != nil {
			return err
		}

		code, err := Rack(c).ProcessRun(app.Name, types.ProcessRunOptions{
			Command:     s.Test,
			Environment: senv,
			Release:     build.Release,
			Service:     s.Name,
			Output:      w,
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
