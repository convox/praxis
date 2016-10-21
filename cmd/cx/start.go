package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/convox/praxis/cli"
	"github.com/convox/praxis/client"
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
	go handleSignals(c)

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	app := filepath.Base(wd)

	a, err := rack().AppCreate(app, client.AppCreateOptions{})
	if err != nil {
		return err
	}

	defer rack().AppDelete(a.Name)

	build, err := buildDirectory(app, ".")
	if err != nil {
		return err
	}

	if err := streamBuild(app, build.Id, os.Stdout); err != nil {
		return err
	}

	build, err = waitForBuild(a.Name, build.Id)
	if err != nil {
		return err
	}

	fmt.Printf("build = %+v\n", build)

	// Provider.ReleasePromote(build.Release)
	// Provider.ReleaseWait(build.Release)
	// Provider.ServiceScale("web", {Count:1})

	// SIGINT ->
	// ps := range Provider.ProcessList()
	//   Provider.ProcessStop(ps)

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
