package main

import (
	"io"
	"os"
	"time"

	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "logs",
		Description: "show app logs",
		Action:      runLogs,
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
	})
}

func runLogs(c *cli.Context) error {
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

	logs, err := Rack.AppLogs(app, opts)
	if err != nil {
		return err
	}

	if _, err := io.Copy(os.Stdout, logs); err != nil {
		return err
	}

	return nil
}
