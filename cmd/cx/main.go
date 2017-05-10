package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	homedir "github.com/mitchellh/go-homedir"
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

	stdcli.DefaultWriter.Tags["app"] = stdcli.RenderAttributes(39)
	// stdcli.DefaultWriter.Tags["config"] = stdcli.RenderAttributes()
	// stdcli.DefaultWriter.Tags["description"] = stdcli.RenderAttributes(245)
	// stdcli.DefaultWriter.Tags["fail"] = stdcli.RenderAttributes(160)
	// stdcli.DefaultWriter.Tags["file"] = stdcli.RenderAttributes(249)
	// stdcli.DefaultWriter.Tags["link"] = stdcli.RenderAttributes(4)
	// stdcli.DefaultWriter.Tags["linenumber"] = stdcli.RenderAttributes(235)
	// stdcli.DefaultWriter.Tags["release"] = stdcli.RenderAttributes(39)
	// stdcli.DefaultWriter.Tags["security"] = stdcli.RenderAttributes(226)
	// stdcli.DefaultWriter.Tags["service"] = stdcli.RenderAttributes(39)
	// stdcli.DefaultWriter.Tags["database"] = stdcli.RenderAttributes(141)
	// stdcli.DefaultWriter.Tags["success"] = stdcli.RenderAttributes(10)
	// stdcli.DefaultWriter.Tags["unsupported"] = stdcli.RenderAttributes(220)
	// stdcli.DefaultWriter.Tags["warning"] = stdcli.RenderAttributes(172)
}

func main() {
	app := stdcli.New()

	app.Name = "cx"
	app.Version = Version
	app.Usage = "convox management tool"

	stdcli.VersionPrinter(func(c *cli.Context) {
		runVersion(c)
	})

	ch := make(chan error)

	go autoUpdate(ch)

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err := <-ch; err != nil {
		// fmt.Fprintf(os.Stderr, "error: autoupdate: %v\n", err)
		// os.Exit(1)
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

func autoUpdate(ch chan error) {
	var updated time.Time

	home, err := homedir.Dir()
	if err != nil {
		ch <- err
		return
	}

	setting := filepath.Join(home, ".convox", "updated")

	if data, err := ioutil.ReadFile(setting); err == nil {
		up, err := time.Parse(helpers.SortableTime, string(data))
		if err != nil {
			ch <- err
			return
		}
		updated = up
	}

	if updated.After(time.Now().UTC().Add(-1 * time.Hour)) {
		time.Sleep(1 * time.Second)
		ch <- nil
		return
	}

	os.MkdirAll(filepath.Dir(setting), 0755)
	ioutil.WriteFile(setting, []byte(time.Now().UTC().Format(helpers.SortableTime)), 0644)

	ex, err := os.Executable()
	if err != nil {
		ch <- err
		return
	}

	v, err := latestVersion()
	if err != nil {
		ch <- err
		return
	}

	exec.Command(ex, "update", v).Start()

	ch <- nil
}

func errorExit(fn cli.ActionFunc, code int) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := fn(c); err != nil {
			return cli.NewExitError(err.Error(), code)
		}
		return nil
	}
}
