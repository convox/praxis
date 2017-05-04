package main

import (
	"os"

	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	shellquote "github.com/kballard/go-shellquote"
	"gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "run",
		Description: "run a new process",
		Usage:       "<service> [command]",
		Action:      runRun,
		Flags: []cli.Flag{
			appFlag,
		},
	})
}

func runRun(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	if len(c.Args()) < 1 {
		return stdcli.Usage(c)
	}

	service := c.Args()[0]
	command := ""

	if len(c.Args()) >= 2 {
		command = shellquote.Join(c.Args()[1:]...)
	}

	opts := types.ProcessRunOptions{
		Command: command,
		Service: service,
		Stream:  types.Stream{Reader: os.Stdin, Writer: os.Stdout},
	}

	code, err := Rack.ProcessRun(app, opts)
	if err != nil {
		return err
	}

	os.Exit(code)

	return nil
}
