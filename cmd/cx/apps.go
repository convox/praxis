package main

import (
	"fmt"

	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "apps",
		Description: "list applications",
		Action:      runApps,
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "create",
				Description: "create an application",
				Usage:       "<name>",
				Action:      runAppsCreate,
			},
			cli.Command{
				Name:        "delete",
				Description: "delete an application",
				Usage:       "<name>",
				Action:      runAppsDelete,
			},
			cli.Command{
				Name:        "info",
				Description: "get info about an application",
				Usage:       "<name>",
				Action:      runAppsInfo,
			},
		},
	})
}

func runApps(c *cli.Context) error {
	apps, err := Rack.AppList()
	if err != nil {
		return err
	}

	t := stdcli.NewTable("NAME", "STATUS")

	for _, app := range apps {
		t.AddRow(app.Name, app.Status)
	}

	t.Print()

	return nil
}

func runAppsCreate(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	app, err := Rack.AppCreate(c.Args()[0])
	if err != nil {
		return err
	}

	fmt.Printf("app = %+v\n", app)

	return nil
}

func runAppsDelete(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	if err := Rack.AppDelete(c.Args()[0]); err != nil {
		return err
	}

	return nil
}

func runAppsInfo(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	app := c.Args()[0]

	a, err := Rack.AppGet(app)
	if err != nil {
		return err
	}

	info := stdcli.NewInfo()

	info.Add("Name", a.Name)
	info.Add("Release", a.Release)
	info.Add("Status", a.Status)

	info.Print()

	return nil
}
