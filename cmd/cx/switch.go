package main

import (
	"os"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "switch",
		Description: "switch to another rack",
		Usage:       "[RACK]",
		Action:      runSwitch,
	})
}

func runSwitch(c *cli.Context) error {
	if len(c.Args()) == 0 {
		sw := *stdcli.DefaultWriter

		if os.Getenv("RACK_URL") != "" {
			r, err := rack.NewFromEnv()
			if err != nil {
				return stdcli.Error(err)
			}

			s, err := r.SystemGet()
			if err != nil {
				return stdcli.Error(err)
			}

			sw.Writef("RACK_URL/%s\n", s.Name)
			return nil
		}

		rack, err := currentRack(c)
		if err != nil {
			return stdcli.Error(err)
		}

		sw.Writef("%s\n", rack)
		return nil
	}

	sr := c.Args()[0]

	if sr == "local" {
		return setShellRack("local")
	}

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
		return stdcli.Errorf("Rack %q was not found", sr)
	}

	return setShellRack(sr)
}
