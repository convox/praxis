package main

import (
	"os"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "test",
		Description: "run tests",
		Action:      runTest,
	})
}

func runTest(c *cli.Context) error {
	m, err := manifest.LoadFile("convox.yml")
	if err != nil {
		return err
	}

	err = m.Build("test", manifest.BuildOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return err
	}

	err = m.Test("test", manifest.TestOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return err
	}

	return nil
}
