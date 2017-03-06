package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	"gopkg.in/urfave/cli.v1"
)

var (
	Rack    rack.Rack
	Version = "dev"
)

var (
	appFlag = cli.StringFlag{
		Name:  "app, a",
		Usage: "app name",
	}
)

func init() {
	host := "localhost:5443"

	if rh := os.Getenv("RACK_HOST"); rh != "" {
		host = rh
	}

	Rack = rack.New(host)
}

func main() {
	// os.Remove("/tmp/test.sock")
	// go server.New().Listen("unix", "/tmp/test.sock")
	// time.Sleep(100 * time.Millisecond)
	// Rack.Socket = "/tmp/test.sock"

	app := stdcli.New()

	app.Name = "cx"
	app.Version = Version
	app.Usage = "convox management tool"

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
}
