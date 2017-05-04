package main

import (
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "resources",
		Description: "list resources",
		Action:      runResources,
		Flags:       []cli.Flag{appFlag},
	})
}

func runResources(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	rs, err := Rack.ResourceList(app)
	if err != nil {
		return err
	}

	t := stdcli.NewTable("NAME", "TYPE", "ENDPOINT")

	for _, r := range rs {
		t.AddRow(r.Name, r.Type, r.Endpoint)
	}

	t.Print()

	return nil
}
