package main

import (
	"os"

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
		Flags:       globalFlags,
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

	state, err := terminalRaw(os.Stdin)
	if err != nil {
		return err
	}
	defer terminalRestore(os.Stdin, state)

	code, err := Rack(c).ProcessExec(app, pid, command, types.ProcessExecOptions{
		Input:  os.Stdin,
		Output: os.Stdout,
	})
	if err != nil {
		return err
	}

	terminalRestore(os.Stdin, state)

	os.Exit(code)

	return nil
}
