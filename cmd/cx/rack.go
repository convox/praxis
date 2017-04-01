package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/convox/praxis/frontend"
	"github.com/convox/praxis/provider"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
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
				Usage:       "<provider> <name>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "version",
						Usage: "rack version",
						Value: "latest",
					},
				},
			},
			cli.Command{
				Name:        "start",
				Description: "start a local rack",
				Action:      runRackStart,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "frontend",
						Usage: "frontend host",
						Value: "10.42.84.0",
					},
				},
			},
			cli.Command{
				Name:        "uninstall",
				Description: "uninstall a rack",
				Action:      runRackUninstall,
				Usage:       "<provider> <name>",
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
	if len(c.Args()) != 2 {
		return stdcli.Usage(c)
	}

	ptype := c.Args()[0]
	name := c.Args()[1]

	key, err := types.Key(32)
	if err != nil {
		return err
	}

	switch ptype {
	case "aws":
		if err := fetchCredentialsAWS(); err != nil {
			return err
		}
	}

	p, err := provider.FromType(ptype)
	if err != nil {
		return err
	}

	endpoint, err := p.SystemInstall(name, types.SystemInstallOptions{
		Color:   true,
		Key:     key,
		Output:  os.Stdout,
		Version: c.String("version"),
	})
	if err != nil {
		return err
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	u.User = url.UserPassword(key, "")

	fmt.Printf("RACK_URL=%s\n", u.String())

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

	cmd, err := rackCommand(version, c.String("frontend"))
	if err != nil {
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func runRackUninstall(c *cli.Context) error {
	if len(c.Args()) != 2 {
		return stdcli.Usage(c)
	}

	ptype := c.Args()[0]
	name := c.Args()[1]

	switch ptype {
	case "aws":
		if err := fetchCredentialsAWS(); err != nil {
			return err
		}
	}

	p, err := provider.FromType(ptype)
	if err != nil {
		return err
	}

	err = p.SystemUninstall(name, types.SystemInstallOptions{
		Color:  true,
		Output: os.Stdout,
	})
	if err != nil {
		return err
	}

	return nil
}

func runRackUpdate(c *cli.Context) error {
	return nil
}

func rackCommand(version string, frontend string) (*exec.Cmd, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	exec.Command("docker", "rm", "-f", "rack").Run()

	args := []string{"run"}
	args = append(args, "-e", "PROVIDER=local")
	args = append(args, "-e", fmt.Sprintf("PROVIDER_FRONTEND=%s", frontend))
	args = append(args, "-e", fmt.Sprintf("VERSION=%s", version))
	args = append(args, "-i", "--rm", "--name=rack")
	args = append(args, "-p", "5443:3000")
	args = append(args, "-v", fmt.Sprintf("%s:/var/convox", filepath.Join(home, ".convox", "local")))
	args = append(args, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	args = append(args, fmt.Sprintf("convox/praxis:%s", version))

	return exec.Command("docker", args...), nil
}

func aws(args ...string) ([]byte, error) {
	var buf bytes.Buffer

	cmd := exec.Command("aws", args...)

	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func fetchCredentialsAWS() error {
	data, err := aws("configure", "get", "region")
	if err != nil || len(data) == 0 {
		return fmt.Errorf("aws cli must be configured, try `aws configure`")
	}

	os.Setenv("AWS_REGION", strings.TrimSpace(string(data)))

	data, err = aws("configure", "get", "role_arn")
	if err == nil && len(data) > 0 {
		return fetchCredentialsAWSRole(strings.TrimSpace(string(data)))
	}

	data, err = aws("configure", "get", "aws_access_key_id")
	if err != nil || len(data) == 0 {
		return fmt.Errorf("aws cli must be configured, try `aws configure`")
	}

	os.Setenv("AWS_ACCESS_KEY_ID", strings.TrimSpace(string(data)))

	data, err = aws("configure", "get", "aws_secret_access_key")
	if err != nil || len(data) == 0 {
		return fmt.Errorf("aws cli must be configured, try `aws configure`")
	}

	os.Setenv("AWS_SECRET_ACCESS_KEY", strings.TrimSpace(string(data)))

	return nil
}

func fetchCredentialsAWSRole(role string) error {
	data, err := aws("sts", "assume-role", "--role-arn", role, "--role-session-name", "convox-cli")
	if err != nil {
		return err
	}

	var auth struct {
		Credentials struct {
			AccessKeyId     string
			SecretAccessKey string
			SessionToken    string
		}
	}

	if err := json.Unmarshal(data, &auth); err != nil {
		return err
	}

	os.Setenv("AWS_ACCESS_KEY_ID", auth.Credentials.AccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", auth.Credentials.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", auth.Credentials.SessionToken)

	return nil
}
