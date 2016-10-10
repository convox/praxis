package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/cli"
	"github.com/convox/praxis/manifest"
)

func init() {
	cli.Register(cli.Command{
		Func:    cmdStart,
		Name:    "start",
		Usage:   "[process] [command...]",
		Summary: "start a convox app locally",
		Description: `
			lorem ipsum dolor
			  sit amet

			consectetur adipiscing elit
		`,
	})
}

func main() {
	cli.Init(cli.Settings{
		Name:    "cx",
		Summary: "build, run, and deploy convox apps",
		Description: `
			convox is a command

			line
			  application
		`,
	})

	cli.Run(os.Args)
}

func cmdStart(c cli.Context) error {
	m, err := manifest.LoadFile("convox.yml")
	if err != nil {
		return err
	}

	data, err := m.Raw()
	if err != nil {
		return err
	}

	fmt.Printf("m = %+v\n", m)
	fmt.Printf("string(data) = \n%+v\n", string(data))

	if err := m.Build(); err != nil {
		return err
	}

	if err := m.Run(); err != nil {
		return err
	}

	return nil
}
