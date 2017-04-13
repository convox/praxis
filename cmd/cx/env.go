package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "env",
		Description: "display current env",
		Action:      runEnv,
		Flags: []cli.Flag{
			appFlag,
		},
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "set",
				Description: "change env values",
				Usage:       "<KEY=value> [KEY=value]...",
				Action:      runEnvSet,
				Flags: []cli.Flag{
					appFlag,
				},
			},
			cli.Command{
				Name:        "unset",
				Description: "remove env values",
				Usage:       "<KEY> [KEY]...",
				Action:      runEnvUnset,
				Flags: []cli.Flag{
					appFlag,
				},
			},
		},
	})
}

func runEnv(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	env, err := Rack.EnvironmentGet(app)
	if err != nil {
		return err
	}

	keys := []string{}

	for k := range env {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("%s=%s\n", k, env[k])
	}

	return nil
}

func runEnvSet(c *cli.Context) error {
	env := types.Environment{}

	if !stdcli.IsTerminal(os.Stdin) {
		env.Read(os.Stdin)
	} else {
		if len(c.Args()) < 1 {
			return stdcli.Usage(c)
		}
	}

	env.Pairs(c.Args())

	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	if err := Rack.EnvironmentSet(app, env); err != nil {
		return err
	}

	return nil
}

func runEnvUnset(c *cli.Context) error {
	if len(c.Args()) < 1 {
		return stdcli.Usage(c)
	}

	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	for _, key := range c.Args() {
		if err := Rack.EnvironmentUnset(app, key); err != nil {
			return err
		}
	}

	return nil
}
