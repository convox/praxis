package main

import (
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "info",
		Description: "get application info",
		Action:      runAppsInfo,
		Flags:       globalFlags,
	})
}
