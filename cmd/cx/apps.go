package main

import (
	"fmt"
	"sort"

	"github.com/convox/praxis/manifest"
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

	if a.Release != "" {
		r, err := Rack.ReleaseGet(app, a.Release)
		if err != nil {
			return err
		}

		if r.Build != "" {
			sys, err := Rack.SystemGet()
			if err != nil {
				return err
			}

			b, err := Rack.BuildGet(app, r.Build)
			if err != nil {
				return err
			}

			m, err := manifest.Load([]byte(b.Manifest))
			if err != nil {
				return err
			}

			endpoints := []string{}

			for _, s := range m.Services {
				if s.Port.Port > 0 {
					endpoints = append(endpoints, fmt.Sprintf("https://%s-%s.%s", app, s.Name, sys.Domain))
				}
			}

			sort.Strings(endpoints)

			info.Add("Endpoints", endpoints...)
		}
	}

	info.Print()

	return nil
}
