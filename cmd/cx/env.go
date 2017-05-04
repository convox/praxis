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

	rs, err := Rack.ReleaseList(app, types.ReleaseListOptions{Count: 1})
	if err != nil {
		return err
	}

	if len(rs) < 1 {
		return nil
	}

	env := rs[0].Env

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

	cenv := types.Environment{}

	rs, err := Rack.ReleaseList(app, types.ReleaseListOptions{Count: 1})
	if err != nil {
		return err
	}

	if len(rs) > 0 {
		cenv = rs[0].Env
	}

	for k, v := range env {
		cenv[k] = v
	}

	if err := releaseCreate(app, types.ReleaseCreateOptions{Env: cenv}); err != nil {
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

	cenv := types.Environment{}

	rs, err := Rack.ReleaseList(app, types.ReleaseListOptions{Count: 1})
	if err != nil {
		return err
	}

	if len(rs) > 0 {
		cenv = rs[0].Env
	}

	for _, k := range c.Args() {
		delete(cenv, k)
	}

	if err := releaseCreate(app, types.ReleaseCreateOptions{Env: cenv}); err != nil {
		return err
	}

	return nil
}
