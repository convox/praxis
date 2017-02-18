package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/convox/praxis/stdcli"
	homedir "github.com/mitchellh/go-homedir"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "rack",
		Description: "show system information",
		Action:      runRack,
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "start",
				Description: "start a local rack",
				Action:      runRackStart,
			},
			cli.Command{
				Name:        "update",
				Description: "update the rack",
				Usage:       "[version]",
				Action:      runRackUpdate,
			},
		},
	})
}

func runRack(c *cli.Context) error {
	rack, err := Rack.SystemGet()
	if err != nil {
		return err
	}

	fmt.Printf("rack = %+v\n", rack)

	return nil
}

func runRackStart(c *cli.Context) error {
	version := "latest"

	switch len(c.Args()) {
	case 0:
	case 1:
		version = c.Args()[0]
	default:
		return stdcli.Usage(c)
	}

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	exec.Command("docker", "rm", "-f", "rack").Run()

	args := []string{"run"}
	args = append(args, "-i", "--rm", "--name=rack")
	args = append(args, "-p", "5443:3000")
	args = append(args, "-v", fmt.Sprintf("%s:/var/convox", filepath.Join(home, ".convox", "local")))
	args = append(args, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	args = append(args, fmt.Sprintf("convox/praxis:%s", version))

	cmd := exec.Command("docker", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func runRackUpdate(c *cli.Context) error {
	return nil
}
