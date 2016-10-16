package main

import (
	"github.com/convox/praxis/api"
	"github.com/convox/praxis/cli"
)

func init() {
	cli.Register(cli.Command{
		Func:    cmdDaemon,
		Name:    "daemon",
		Usage:   "",
		Summary: "start a local rack",
	})
}

func cmdDaemon(c cli.Context) error {
	if err := api.Listen("localhost:9877"); err != nil {
		return err
	}

	return nil
}
