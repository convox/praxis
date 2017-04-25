package main

import (
	"fmt"
	"math/rand"

	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "resources",
		Description: "list resources",
		Action:      runResources,
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "create",
				Description: "create an resource",
				Usage:       "<type>",
				Action:      runResourcesCreate,
			},
			cli.Command{
				Name:        "delete",
				Description: "delete an resource",
				Usage:       "<name>",
				Action:      runResourcesDelete,
			},
			cli.Command{
				Name:        "info",
				Description: "get info about an resource",
				Usage:       "<name>",
				Action:      runResourcesInfo,
			},
		},
	})
}

func runResources(c *cli.Context) error {
	ress, err := Rack.ResourceList()
	if err != nil {
		return err
	}

	t := stdcli.NewTable("NAME", "TYPE", "STATUS")

	for _, r := range ress {
		t.AddRow(r.Name, r.Type, r.Status)
	}

	t.Print()

	return nil
}

func runResourcesCreate(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	kind := c.Args()[0]
	name := fmt.Sprintf("%s-%d", kind, (rand.Intn(8999) + 1000))

	resource, err := Rack.ResourceCreate(kind, name, nil)
	if err != nil {
		return err
	}

	fmt.Printf("resource = %+v\n", resource)

	return nil
}

func runResourcesDelete(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	if err := Rack.AppDelete(c.Args()[0]); err != nil {
		return err
	}

	return nil
}

func runResourcesInfo(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	res, err := Rack.ResourceGet(c.Args()[0])
	if err != nil {
		return err
	}

	info := stdcli.NewInfo()

	info.Add("Name", res.Name)
	info.Add("Type", res.Type)
	info.Add("Status", res.Status)

	info.Print()

	return nil
}
