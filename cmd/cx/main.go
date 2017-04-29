package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	r, err := rack.NewFromEnv()
	if err != nil {
		panic(err)
	}

	Rack = r
}

func main() {
	app := stdcli.New()

	app.Name = "cx"
	app.Version = Version
	app.Usage = "convox management tool"

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func appName(c *cli.Context, dir string) (string, error) {
	if app := c.String("app"); app != "" {
		return app, nil
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	return filepath.Base(abs), nil
}

func errorExit(fn cli.ActionFunc, code int) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := fn(c); err != nil {
			return cli.NewExitError(err.Error(), code)
		}
		return nil
	}
}
