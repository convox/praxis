package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "filter",
			Usage: "filter logs",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "follow, f",
			Usage: "stream logs continuously",
		},
	}
	stdcli.RegisterCommand(cli.Command{
		Name:        "releases",
		Description: "list releases",
		Action:      runReleases,
		Flags:       globalFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "info",
				Description: "release info",
				Usage:       "<id>",
				Action:      runReleasesInfo,
				Flags:       globalFlags,
			},
			cli.Command{
				Name:        "logs",
				Description: "release logs",
				Usage:       "<id>",
				Action:      runReleasesLogs,
				Flags:       append(flags, globalFlags...),
			},
		},
	})
}

func runReleases(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	releases, err := Rack(c).ReleaseList(app, types.ReleaseListOptions{Count: 10})
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

	r, err := Rack(c).ReleaseGet(app, id)
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

	r, err := Rack(c).ReleaseGet(app, id)
	if err != nil {
		return err
	}

	opts := types.LogsOptions{
		Filter: c.String("filter"),
		Follow: c.Bool("follow"),
		Since:  r.Created,
	}

	fmt.Printf("opts = %+v\n", opts)

	err = releaseLogs(Rack(c), app, id, os.Stdout, opts)
	return err
}

func notReleaseStatus(r rack.Rack, app, id, status string) func() (bool, error) {
	return func() (bool, error) {
		rs, err := r.ReleaseGet(app, id)
		if err != nil {
			return true, err
		}
		if rs.Status != status {
			return true, nil
		}

		return false, nil
	}
}

type progress int64

func (p *progress) Write(data []byte) (int, error) {
	*p += progress(len(data))
	return len(data), nil
}

func releaseLogs(r rack.Rack, app string, id string, w io.Writer, opts types.LogsOptions) error {
	if err := tickWithTimeout(2*time.Second, 5*time.Minute, notReleaseStatus(r, app, id, "created")); err != nil {
		return err
	}

	var p progress

	for {
		logs, err := r.ReleaseLogs(app, id, opts)
		if err != nil {
			return err
		}

		if _, err := io.CopyN(ioutil.Discard, logs, int64(p)); err != nil {
			return err
		}

		if _, err := io.Copy(io.MultiWriter(w, &p), logs); err != nil {
			return err
		}

		r, err := r.ReleaseGet(app, id)
		if err != nil {
			return err
		}

		switch r.Status {
		case "promoted", "failed":
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}
