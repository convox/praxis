package main

import (
	"fmt"

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

	fmt.Printf("rs = %+v\n", rs)
	// t := stdcli.NewTable("NAME", "ENDPOINT")

	// for _, s := range ss {
	//   t.AddRow(s.Name, s.Endpoint)
	// }

	// t.Print()

	return nil
}
