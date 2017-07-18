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
	fmt.Printf("client: %s\n", Version)

	rack, err := Rack(c).SystemGet()
	if err != nil {
		fmt.Printf("server: error: %s\n", err)
		return err
	}

	fmt.Printf("server: %s\n", rack.Version)

	return nil
}

func latestVersion(channel string) (string, error) {
	fmt.Printf("channel = %+v\n", channel)

	req, err := http.NewRequest("GET", fmt.Sprintf("https://releases.convox.com/releases/%s/next", channel), nil)
	if err != nil {
		return "", err
	}

	agent := fmt.Sprintf("convox/%s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH)

	if id, _ := cliID(); id != "" {
		agent = fmt.Sprintf("%s (%s)", agent, id)
	}

	req.Header.Set("User-Agent", agent)

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
