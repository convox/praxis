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
		Flags:       globalFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "set",
				Description: "change env values",
				Usage:       "<KEY=value> [KEY=value]...",
				Action:      runEnvSet,
				Flags:       globalFlags,
			},
			cli.Command{
				Name:        "unset",
				Description: "remove env values",
				Usage:       "<KEY> [KEY]...",
				Action:      runEnvUnset,
				Flags:       globalFlags,
			},
		},
	})
}

func runEnv(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	rs, err := Rack(c).ReleaseList(app, types.ReleaseListOptions{Count: 1})
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

	rs, err := Rack(c).ReleaseList(app, types.ReleaseListOptions{Count: 1})
	if err != nil {
		return err
	}

	if len(rs) > 0 {
		cenv = rs[0].Env
	}

	for k, v := range env {
		cenv[k] = v
	}

	stdcli.Startf("updating environment")

	_, err = Rack(c).ReleaseCreate(app, types.ReleaseCreateOptions{Env: cenv})
	if err != nil {
		return stdcli.Error(err)
	}

	stdcli.OK()

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

	rs, err := Rack(c).ReleaseList(app, types.ReleaseListOptions{Count: 1})
	if err != nil {
		return err
	}

	if len(rs) > 0 {
		cenv = rs[0].Env
	}

	for _, k := range c.Args() {
		delete(cenv, k)
	}

	stdcli.Startf("updating environment")

	_, err = Rack(c).ReleaseCreate(app, types.ReleaseCreateOptions{Env: cenv})
	if err != nil {
		return stdcli.Error(err)
	}

	stdcli.OK()

	return nil
}
