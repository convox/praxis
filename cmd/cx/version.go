package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "version",
		Description: "display cli version",
		Action:      runVersion,
	})
}

func runVersion(c *cli.Context) error {
	rack, err := Rack.SystemGet()
	if err != nil {
		return err
	}

	fmt.Printf("client: %s\n", Version)
	fmt.Printf("server: %s\n", rack.Version)

	return nil
}

func latestVersion() (string, error) {
	req, err := http.NewRequest("GET", "https://releases.convox.com/releases/edge/next", nil)
	if err != nil {
		return "", err
	}

	id, err := cliID()
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("convox/%s (%s; %s/%s)", Version, id, runtime.GOOS, runtime.GOARCH))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var next string

	if err := json.Unmarshal(data, &next); err != nil {
		return "", err
	}

	return next, nil
}
