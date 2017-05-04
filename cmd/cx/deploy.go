package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "deploy",
		Description: "deploy the application",
		Action:      runDeploy,
		Flags: []cli.Flag{
			appFlag,
		},
	})
}

func runDeploy(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	a, err := Rack.AppGet(app)
	if err != nil {
		return err
	}

	if a.Status != "running" {
		return fmt.Errorf("cannot build while app is %s", a.Status)
	}

	build, err := buildDirectory(app, ".", os.Stdout)
	if err != nil {
		return err
	}

	if err := buildLogs(build, os.Stdout); err != nil {
		return err
	}

	build, err = Rack.BuildGet(app, build.Id)
	if err != nil {
		return err
	}

	if build.Status == "failed" {
		return fmt.Errorf("build failed")
	}

	if err := releaseLogs(app, build.Release, os.Stdout); err != nil {
		return err
	}

	r, err := Rack.ReleaseGet(app, build.Release)
	if err != nil {
		return err
	}

	if r.Status != "complete" {
		return fmt.Errorf("deploy failed")
	}

	return nil
}
