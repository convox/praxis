package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"

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

var currentManifest *manifest.Manifest

func cmdStart(c cli.Context) error {
	m, err := manifest.LoadFile("convox.yml")
	if err != nil {
		return err
	}

	u, err := user.Current()
	if err != nil {
		return err
	}

	for i := range m.Services {
		for j := range m.Services[i].Volumes {
			m.Services[i].Volumes[j].Local = filepath.Join(u.HomeDir, ".convox", "volumes", m.Services[i].Volumes[j].Local)
		}
	}

	go handleSignals(c)

	build, err := buildDirectory(".")
	if err != nil {
		return err
	}

	fmt.Printf("build = %+v\n", build)

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