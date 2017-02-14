package main

import (
	"fmt"
	"os"
	"time"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/server"
	"github.com/convox/praxis/stdcli"
)

var (
	Rack    *rack.Client
	Version = "dev"
)

func init() {
	Rack = rack.New()
}

func main() {
	os.Remove("/tmp/test.sock")
	go server.New().Listen("unix", "/tmp/test.sock")
	time.Sleep(100 * time.Millisecond)
	Rack.Socket = "/tmp/test.sock"

	app := stdcli.New()

	app.Name = "cx"
	app.Version = Version
	app.Usage = "convox management tool"

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
}
