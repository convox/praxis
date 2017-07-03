package main

import (
	"time"

	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "apps",
		Description: "list applications",
		Action:      runApps,
		Before:      beforeCmd,
		Flags:       globalFlags,
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "create",
				Description: "create an application",
				Usage:       "<name>",
				Action:      runAppsCreate,
				Flags:       globalFlags,
			},
			cli.Command{
				Name:        "delete",
				Aliases:     []string{"rm"},
				Description: "delete an application",
				Usage:       "<name>",
				Action:      runAppsDelete,
				Flags:       globalFlags,
			},
			cli.Command{
				Name:        "info",
				Description: "get application info",
				Usage:       "[name]",
				Action:      runAppsInfo,
				Flags:       globalFlags,
			},
		},
	})
}

func runApps(c *cli.Context) error {
	apps, err := Rack.AppList()
	if err != nil {
		return stdcli.Error(err)
	}

	t := stdcli.NewTable("NAME", "STATUS")

	for _, app := range apps {
		t.AddRow(app.Name, app.Status)
	}

	t.Print()

	return nil
}

func runAppsCreate(c *cli.Context) error {
	name, err := appName(c, ".")
	if err != nil {
		return stdcli.Error(err)
	}

	if len(c.Args()) > 0 {
		name = c.Args()[0]
	}

	stdcli.Startf("creating <name>%s</name>", name)

	if _, err = Rack.AppCreate(name); err != nil {
		return stdcli.Error(err)
	}

	if err := tickWithTimeout(2*time.Second, 1*time.Minute, notAppStatus(name, "creating")); err != nil {
		return err
	}

	stdcli.OK()

	return nil
}

func runAppsDelete(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	name := c.Args()[0]

	stdcli.Startf("deleting <name>%s</name>", name)

	if err := Rack.AppDelete(name); err != nil {
		return stdcli.Error(err)
	}

	stdcli.OK()

	return nil
}

func runAppsInfo(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return stdcli.Error(err)
	}

	if len(c.Args()) > 0 {
		app = c.Args()[0]
	}

	a, err := Rack.AppGet(app)
	if err != nil {
		return stdcli.Error(err)
	}

	info := stdcli.NewInfo()

	info.Add("Name", a.Name)
	info.Add("Release", a.Release)
	info.Add("Status", a.Status)

	info.Print()

	return nil
}

func notAppStatus(app, status string) func() (bool, error) {
	return func() (bool, error) {
		app, err := Rack.AppGet(app)
		if err != nil {
			return true, err
		}
		if app.Status != status {
			return true, nil
		}

		return false, nil
	}
}
