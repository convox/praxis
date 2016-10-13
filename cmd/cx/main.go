package main

import (
	"os"
	"os/signal"

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

var currentManifest *manifest.Manifest

func cmdStart(c cli.Context) error {
	m, err := manifest.LoadFile("convox.yml")
	if err != nil {
		return err
	}

	go handleSignals(c)

	if err := m.Build(); err != nil {
		return err
	}

	currentManifest = m

	if err := m.Run(manifest.RunOptions{Sync: true}); err != nil {
		return err
	}

	return nil
}

func handleSignals(c cli.Context) {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)

	for sig := range ch {
		switch sig {
		case os.Interrupt:
			c.Printf("\n")

			if currentManifest != nil {
				currentManifest.Stop()
			}

			os.Exit(1)
		}
	}
}
