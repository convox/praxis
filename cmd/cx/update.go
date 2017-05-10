package main

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/convox/praxis/stdcli"
	update "github.com/inconshreveable/go-update"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "update",
		Description: "update the cli",
		Usage:       "<version>",
		Action:      runUpdate,
	})
}

func runUpdate(c *cli.Context) error {
	version, err := latestVersion()
	if err != nil {
		return err
	}

	if len(c.Args()) > 0 {
		version = c.Args()[0]
	}

	url := fmt.Sprintf("https://s3.amazonaws.com/praxis-releases/release/%s/cli/%s/cx", version, runtime.GOOS)

	stdcli.Startf("updating cli to <version>%s</version>", version)

	res, err := http.Get(url)
	if err != nil {
		return stdcli.Error(err)
	}

	defer res.Body.Close()

	if err := update.Apply(res.Body, update.Options{}); err != nil {
		return stdcli.Error(err)
	}

	stdcli.OK()

	return nil
}
