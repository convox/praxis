package main

import (
	"fmt"

	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "services",
		Description: "list services",
		Action:      runServices,
		Before:      beforeCmd,
		Flags:       globalFlags,
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "url",
				Description: "output url for a service",
				Usage:       "<name>",
				Action:      runServicesUrl,
				Flags:       globalFlags,
			},
		},
	})
}

func runServices(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	ss, err := Rack(c).ServiceList(app)
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

func runServicesUrl(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	if len(c.Args()) < 1 {
		return stdcli.Usage(c)
	}

	name := c.Args()[0]

	s, err := Rack(c).ServiceGet(app, name)
	if err != nil {
		return err
	}

	fmt.Println(s.Endpoint)

	return nil
}
