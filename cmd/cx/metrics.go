package main

import (
	"fmt"

	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "metrics",
		Description: "show app metrics",
		Flags: []cli.Flag{
			appFlag,
		},
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "list",
				Description: "list metrics",
				Usage:       "<namespace>",
				Action:      listMetrics,
			},
			cli.Command{
				Name:        "get",
				Description: "display metrics",
				Usage:       "<namespace> <metric>",
				Action:      getMetrics,
			},
		},
	})
}

func listMetrics(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	if c.NArg() != 1 {
		return stdcli.Usage(c)
	}

	ns := c.Args()[0]

	metrics, err := Rack.MetricList(app, ns, types.MetricListOptions{})
	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		fmt.Printf("No metrics for %s\n", ns)
		return nil
	}

	for _, m := range metrics {
		fmt.Println(m)
	}

	return nil
}

func getMetrics(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	if c.NArg() != 2 {
		return stdcli.Usage(c)
	}

	ns := c.Args()[0]
	m := c.Args()[1]

	metrics, err := Rack.MetricGet(app, ns, m, types.MetricGetOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("METRICS: %+v\n", metrics)
	return nil
}
