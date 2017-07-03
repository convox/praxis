package main

import (
	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	"gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "ps",
		Description: "list processes",
		Action:      runPs,
		Before:      beforeCmd,
		Flags:       globalFlags,
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "stop",
				Description: "stop a process",
				Usage:       "<pid>",
				Action:      runPsStop,
				Flags:       globalFlags,
			},
		},
	})
}

func runPs(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	ps, err := Rack.ProcessList(app, types.ProcessListOptions{})
	if err != nil {
		return err
	}

	t := stdcli.NewTable("ID", "SERVICE", "RELEASE", "STARTED", "COMMAND")

	for _, p := range ps {
		t.AddRow(p.Id, p.Service, p.Release, helpers.HumanizeTime(p.Started), p.Command)
	}

	t.Print()

	return nil
}

func runPsStop(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	if len(c.Args()) < 1 {
		return stdcli.Usage(c)
	}

	pid := c.Args()[0]

	if err := Rack.ProcessStop(app, pid); err != nil {
		return err
	}

	return nil
}
