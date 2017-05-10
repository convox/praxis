package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
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
				Flags: []cli.Flag{
					appFlag,
				},
			},
			cli.Command{
				Name:        "logs",
				Description: "release logs",
				Usage:       "<id>",
				Action:      runReleasesLogs,
				Flags: []cli.Flag{
					appFlag,
					cli.StringFlag{
						Name:  "filter",
						Usage: "filter logs",
						Value: "",
					},
					cli.BoolFlag{
						Name:  "follow, f",
						Usage: "stream logs continuously",
					},
					cli.StringFlag{
						Name:  "since",
						Usage: "how far back to retrieve logs",
						Value: "2m",
					},
				},
			},
		},
	})
}

func runReleases(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	releases, err := Rack.ReleaseList(app, types.ReleaseListOptions{Count: 10})
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

	since, err := time.ParseDuration(c.String("since"))
	if err != nil {
		return err
	}

	opts := types.LogsOptions{
		Filter: c.String("filter"),
		Follow: c.Bool("follow"),
		Prefix: true,
		Since:  time.Now().Add(-1 * since),
	}

	logs, err := Rack.ReleaseLogs(app, id, opts)
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

func releaseCreate(app string, opts types.ReleaseCreateOptions) error {
	r, err := Rack.ReleaseCreate(app, opts)
	if err != nil {
		return err
	}

	if err := releaseLogs(app, r.Id, os.Stdout); err != nil {
		return err
	}

	r, err = Rack.ReleaseGet(app, r.Id)
	if err != nil {
		return err
	}

	if r.Status != "promoted" {
		return fmt.Errorf("release failed")
	}

	return nil
}

func releaseLogs(app string, id string, w io.Writer) error {
	if err := tickWithTimeout(2*time.Second, 5*time.Minute, notReleaseStatus(app, id, "created")); err != nil {
		return err
	}

	logs, err := Rack.ReleaseLogs(app, id, types.LogsOptions{Follow: true})
	if err != nil {
		return err
	}

	if err := helpers.HalfPipe(w, logs); err != nil {
		return err
	}

	return nil
}
