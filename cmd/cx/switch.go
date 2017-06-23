package main

import (
	"net/url"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "switch",
		Description: "swtich to another rack",
		Usage:       "<RACK>",
		Action:      runSwitch,
	})
}

func runSwitch(c *cli.Context) error {
	if len(c.Args()) < 1 {
		return stdcli.Errorf("Please specify a rack to switch")
	}

	proxy, err := consoleProxy()
	if err != nil {
		return stdcli.Error(err)
	}

	if proxy == "" {
		return stdcli.Errorf("Console host not found, try cx login")
	}

	endpoint, err := url.Parse(proxy)
	if err != nil {
		return stdcli.Error(err)
	}

	racks := []string{}
	err = consoleClient(endpoint).Get("/racks", rack.RequestOptions{}, &racks)
	if err != nil {
		return stdcli.Error(err)
	}

	t := stdcli.NewTable("RACKS")

	for _, r := range racks {
		t.AddRow(r)
	}

	t.Print()
	return nil
}
