package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "deploy",
		Description: "build and promote an application",
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

	build, err := buildDirectory(app, ".", types.BuildCreateOptions{})
	if err != nil {
		return err
	}

	if err := Rack.ReleasePromote(app, build.Release); err != nil {
		return err
	}

	if err := releaseLogs(app, build.Release, os.Stdout, types.LogsOptions{Follow: true}); err != nil {
		return err
	}

	r, err := Rack.ReleaseGet(app, build.Release)
	if err != nil {
		return err
	}

	if r.Status != "promoted" {
		return fmt.Errorf("deploy failed")
	}

	return nil
}
