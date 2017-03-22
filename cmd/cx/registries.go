package main

import (
	"fmt"

	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "registries",
		Description: "list registries",
		Action:      runRegistries,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "add",
				Description: "add a registry",
				Action:      runRegistriesAdd,
				Usage:       "<hostname>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "username, u",
						Usage: "registry username",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "registry password",
					},
				},
			},
			cli.Command{
				Name:        "remove",
				Description: "remove a registry",
				Action:      runRegistriesRemove,
				Usage:       "<hostname>",
			},
		},
	})
}

func runRegistries(c *cli.Context) error {
	registries, err := Rack.RegistryList()
	if err != nil {
		return err
	}

	fmt.Printf("registries = %+v\n", registries)

	return nil
}

func runRegistriesAdd(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	hostname := c.Args()[0]
	username := c.String("username")
	password := c.String("password")

	if _, err := Rack.RegistryAdd(hostname, username, password); err != nil {
		return err
	}

	return nil
}

func runRegistriesRemove(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	hostname := c.Args()[0]

	return Rack.RegistryRemove(hostname)
}
