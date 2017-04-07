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

	fmt.Printf("### app:%s\n", a.Name)
	fmt.Printf("Release  %s\n", a.Release)
	fmt.Printf("Status   %s\n", a.Status)

	// if a.Release == "" {
	//   return nil
	// }

	// r, err := Rack.ReleaseGet(app, a.Release)
	// if err != nil {
	//   return err
	// }

	// b, err := Rack.BuildGet(app, r.Build)
	// if err != nil {
	//   return err
	// }

	// m, err := manifest.Load([]byte(b.Manifest))
	// if err != nil {
	//   return err
	// }

	// for _, b := range m.Balancers {
	//   fmt.Println()
	//   fmt.Printf("### balancer:%s\n", b.Name)

	//   t := stdcli.NewTable("SOURCE", "", "TARGET")
	//   t.SkipHeaders = true

	//   for _, e := range b.Endpoints {
	//     source := fmt.Sprintf("%s://%s.%s.convox:%s", e.Protocol, b.Name, a.Name, e.Port)

	//     if e.Redirect != "" {
	//       t.AddRow(source, "->", e.Redirect)
	//     } else {
	//       t.AddRow(source, "=>", e.Target)
	//     }
	//   }

	//   t.Print()
	// }

	// for _, s := range m.Services {
	//   fmt.Println()
	//   fmt.Printf("### service:%s\n", s.Name)
	//   fmt.Printf("Command  %s\n", s.Command)
	//   fmt.Printf("Test     %s\n", s.Test)
	// }

	return nil
}
