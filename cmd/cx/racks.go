package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "racks",
		Description: "list of racks available",
		Action:      runRacks,
	})
}

func runRacks(c *cli.Context) error {
	racks, err := ConsoleProxy().Racks()
	if err != nil {
		return stdcli.Error(err)
	}

	racks = append(racks, "local")

	t := stdcli.NewTable("RACKS")

	for _, r := range racks {
		t.AddRow(r)
	}

	t.Print()

	return nil
}

type ProxyClient struct {
	c *rack.Client
}

func ConsoleProxy() *ProxyClient {
	proxy, err := consoleProxy()
	if err != nil {
		fmt.Fprint(os.Stderr, stdcli.Error(err))
		os.Exit(1)
	}

	if proxy == nil {
		fmt.Fprint(os.Stderr, stdcli.Error(errMissingProxyEndpoint))
		os.Exit(1)
	}

	return &ProxyClient{
		c: &rack.Client{Debug: os.Getenv("CONVOX_DEBUG") == "true", Endpoint: proxy, Version: "dev"},
	}
}

func (p *ProxyClient) Racks() (racks []string, err error) {
	err = p.c.Get("/racks", rack.RequestOptions{}, &racks)
	return
}
