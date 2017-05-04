package main

import (
	"os"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	shellquote "github.com/kballard/go-shellquote"
	"gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "exec",
		Description: "run command inside running process",
		Usage:       "<pid> <command>",
		Action:      runExec,
		Flags: []cli.Flag{
			appFlag,
		},
	})
}

func runExec(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	if len(c.Args()) < 2 {
		return stdcli.Usage(c)
	}

	pid := c.Args()[0]
	command := shellquote.Join(c.Args()[1:]...)

	code, err := Rack.ProcessExec(app, pid, command, types.ProcessExecOptions{Stream: helpers.ReadWriter{Reader: os.Stdin, Writer: os.Stdout}})
	if err != nil {
		return err
	}

	os.Exit(code)

	return nil
}
