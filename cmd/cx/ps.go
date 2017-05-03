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
		Flags: []cli.Flag{
			appFlag,
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
