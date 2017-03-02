package main

import (
	"fmt"

	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	"gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "ps",
		Description: "list processes",
		Action:      runPs,
		Flags: []cli.Flag{
			appFlag,
		},
	})
}

func runPs(c *cli.Context) error {
	ps, err := Rack.ProcessList(c.String("app"), types.ProcessListOptions{})

	fmt.Printf("ps = %+v\n", ps)
	fmt.Printf("err = %+v\n", err)

	return nil
}
