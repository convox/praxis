package main

import (
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "switch",
		Description: "swtich to another rack",
		Usage:       "<RACK>",
		Action:      runSwitch,
	})
}

func runSwitch(c *cli.Context) error {
	if len(c.Args()) < 1 {
		return stdcli.Errorf("Please specify a rack to switch")
	}

	sr := c.Args()[0]

	racks, err := ConsoleProxy().Racks()
	if err != nil {
		return stdcli.Error(err)
	}

	var found bool
	for _, r := range racks {
		if r == sr {
			found = true
			break
		}
	}

	if !found {
		return stdcli.Errorf("Rack %s was not found", sr)
	}

	return setShellRack(sr)
}
