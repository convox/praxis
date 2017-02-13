package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/stdcli"
)

var (
	Version = "dev"
)

func main() {
	app := stdcli.New()

	app.Name = "cx"
	app.Version = Version
	app.Usage = "convox management tool"

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
}
