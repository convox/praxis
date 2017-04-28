package main

import (
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "services",
		Description: "list services",
		Action:      runServices,
		Flags:       []cli.Flag{appFlag},
	})
}

func runServices(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	ss, err := Rack.ServiceList(app)
	if err != nil {
		return err
	}

	t := stdcli.NewTable("NAME", "ENDPOINT")

	for _, s := range ss {
		t.AddRow(s.Name, s.Endpoint)
	}

	t.Print()

	return nil
}
