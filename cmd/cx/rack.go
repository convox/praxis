package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/convox/praxis/frontend"
	"github.com/convox/praxis/sdk/rack"
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
				Name:        "frontend",
				Description: "start a local rack frontend",
				Action:      runRackFrontend,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "interface, i",
						Usage: "interface name",
						Value: "vlan1",
					},
					cli.StringFlag{
						Name:  "subnet, s",
						Usage: "subnet",
						Value: "10.42.84",
					},
				},
			},
			cli.Command{
				Name:        "install",
				Description: "install a rack",
				Action:      runRackInstall,
				Usage:       "<name>",
			},
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

func runRackFrontend(c *cli.Context) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	if u.Uid != "0" {
		return fmt.Errorf("must run as root")
	}

	if err := frontend.Serve(c.String("interface"), c.String("subnet")); err != nil {
		return err
	}

	return nil
}

func runRackInstall(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	name := c.Args()[0]
	version := "test"
	key := "foo"

	template := fmt.Sprintf("https://s3.amazonaws.com/praxis-releases/release/%s/formation/rack.json", version)

	cmd := exec.Command("aws", "cloudformation", "create-stack", "--stack-name", name, "--template-url", template, "--parameters", fmt.Sprintf("ParameterKey=ApiKey,ParameterValue=%s", key), "--capabilities", "CAPABILITY_IAM")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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

	cmd, err := rackCommand(version)
	if err != nil {
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func runRackUpdate(c *cli.Context) error {
	return nil
}

func rackCommand(version string) (*exec.Cmd, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	exec.Command("docker", "rm", "-f", "rack").Run()

	args := []string{"run"}
	args = append(args, "-e", fmt.Sprintf("VERSION=%s", version))
	args = append(args, "-i", "--rm", "--name=rack")
	args = append(args, "-p", "5443:3000")
	args = append(args, "-v", fmt.Sprintf("%s:/var/convox", filepath.Join(home, ".convox", "local")))
	args = append(args, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	args = append(args, fmt.Sprintf("convox/praxis:%s", version))

	return exec.Command("docker", args...), nil
}

func rackRunning() bool {
	data, err := exec.Command("docker", "ps", "--filter", "name=rack", "--format", "{{json .}}").CombinedOutput()
	if err != nil {
		return false
	}

	return len(strings.Split(string(data), "\n")) > 1
}

func startLocalRack() error {
	if rackRunning() {
		return nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	// TODO: make directory, etc
	fd, err := os.OpenFile(filepath.Join(home, ".convox", "rack.log"), os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	cmd, err := rackCommand("latest")
	if err != nil {
		return err
	}

	cmd.Stdout = fd
	cmd.Stderr = fd

	if err := cmd.Start(); err != nil {
		return err
	}

	rk := rack.New("localhost:5443")

	for {
		if _, err := rk.AppList(); err == nil {
			break
		}

		time.Sleep(20 * time.Millisecond)
	}

	return nil
}
