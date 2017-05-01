package main

import (
	"fmt"

	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "version",
		Description: "display cli version",
		Action:      runVersion,
	})
}

func runVersion(c *cli.Context) error {
	rack, err := Rack.SystemGet()
	if err != nil {
		return err
	}

	fmt.Printf("client: %s\n", Version)
	fmt.Printf("server: %s\n", rack.Version)

	return nil
}
