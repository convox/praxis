package main

import (
	"fmt"
	"os"
	"time"

	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "promote",
		Description: "promote a release",
		Action:      runPromote,
		Before:      beforeCmd,
		Flags:       globalFlags,
	})
}

func runPromote(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	rs, err := Rack.ReleaseList(app, types.ReleaseListOptions{Count: 1})
	if err != nil {
		return err
	}

	if len(rs) < 1 {
		return fmt.Errorf("no releases for app: %s", app)
	}

	release := rs[0].Id

	stdcli.Startf("promoting <name>%s</name>", release)

	since := time.Now()

	if err := Rack.ReleasePromote(app, release); err != nil {
		return err
	}

	stdcli.OK()

	if err := releaseLogs(app, release, os.Stdout, types.LogsOptions{Follow: true, Since: since}); err != nil {
		return err
	}

	r, err := Rack.ReleaseGet(app, release)
	if err != nil {
		return err
	}

	if r.Status != "promoted" {
		return fmt.Errorf("promote failed")
	}

	return nil
}
