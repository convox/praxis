package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
	res, err := http.Get("https://api.github.com/repos/convox/praxis/releases?per_page=1")
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var releases []struct {
		Name string
	}

	if err := json.Unmarshal(data, &releases); err != nil {
		return "", err
	}

	if len(releases) < 1 {
		return "", fmt.Errorf("no releases")
	}

	return releases[0].Name, nil
}
