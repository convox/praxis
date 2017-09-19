package main

import (
	"fmt"
	"io/ioutil"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "api",
		Description: "explore the api",
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "get",
				Description: "api GET",
				Usage:       "<path>",
				Action:      runApiGet,
			},
		},
	})
}

func runApiGet(c *cli.Context) error {
	if len(c.Args()) < 1 {
		return stdcli.Usage(c)
	}

	path := c.Args()[0]

	res, err := Rack(c).GetStream(path, rack.RequestOptions{})
	if err != nil {
		return err
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}
