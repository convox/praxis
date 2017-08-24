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
		Flags:       globalFlags,
	})
}

func runPromote(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	rs, err := Rack(c).ReleaseList(app, types.ReleaseListOptions{Count: 1})
	if err != nil {
		return err
	}

	if len(rs) < 1 {
		return fmt.Errorf("no releases for app: %s", app)
	}

	release := rs[0].Id

	stdcli.Startf("promoting <name>%s</name>", release)

	since := time.Now()

	if err := Rack(c).ReleasePromote(app, release); err != nil {
		return err
	}

	stdcli.OK()

	if err := releaseLogs(Rack(c), app, release, os.Stdout, types.LogsOptions{Follow: true, Since: since}); err != nil {
		return err
	}

	r, err := Rack(c).ReleaseGet(app, release)
	if err != nil {
		return err
	}

	switch r.Status {
	case "promoted", "active":
	default:
		return fmt.Errorf("promote failed")
	}

	return nil
}
