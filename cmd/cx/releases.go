package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/convox/praxis/stdcli"
	"github.com/convox/rack/cmd/convox/helpers"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "releases",
		Description: "list releases",
		Action:      runReleases,
		Flags: []cli.Flag{
			appFlag,
		},
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "info",
				Description: "release info",
				Usage:       "<id>",
				Action:      runReleasesInfo,
			},
			cli.Command{
				Name:        "logs",
				Description: "release logs",
				Usage:       "<id>",
				Action:      runReleasesLogs,
			},
		},
	})
}

func runReleases(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	releases, err := Rack.ReleaseList(app)
	if err != nil {
		return err
	}

	t := stdcli.NewTable("ID", "BUILD", "STATUS", "CREATED")

	for _, r := range releases {
		t.AddRow(r.Id, r.Build, r.Status, helpers.HumanizeTime(r.Created))
	}

	t.Print()

	return nil
}

func runReleasesInfo(c *cli.Context) error {
	if len(c.Args()) < 1 {
		stdcli.Usage(c)
	}

	id := c.Args()[0]

	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	r, err := Rack.ReleaseGet(app, id)
	if err != nil {
		return err
	}

	fmt.Printf("r = %+v\n", r)

	return nil
}

func runReleasesLogs(c *cli.Context) error {
	if len(c.Args()) < 1 {
		stdcli.Usage(c)
	}

	id := c.Args()[0]

	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	logs, err := Rack.ReleaseLogs(app, id)
	if err != nil {
		return err
	}

	if _, err := io.Copy(os.Stdout, logs); err != nil {
		return err
	}

	return nil
}

func notReleaseStatus(app, id, status string) func() (bool, error) {
	return func() (bool, error) {
		r, err := Rack.ReleaseGet(app, id)
		if err != nil {
			return true, err
		}
		if r.Status != status {
			return true, nil
		}

		return false, nil
	}
}

func releaseLogs(app string, id string, w io.Writer) error {
	if err := tickWithTimeout(2*time.Second, 5*time.Minute, notReleaseStatus(app, id, "created")); err != nil {
		return err
	}

	logs, err := Rack.ReleaseLogs(app, id)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, logs); err != nil {
		return err
	}

	return nil
}
