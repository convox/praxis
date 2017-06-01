package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/stdcli"
	mv1 "github.com/convox/rack/manifest"
	cli "gopkg.in/urfave/cli.v1"
	yaml "gopkg.in/yaml.v2"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "init",
		Description: "generate convox.yml config",
		Action:      runInit,
	})
}

func runInit(c *cli.Context) error {
	m, err := mv1.LoadFile("docker-compose.yml")
	if err != nil {
		return err
	}

	mNew, err := convert(m)
	if err != nil {
		return err
	}

	ymNew, err := yaml.Marshal(mNew)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("convox.yml", ymNew, 0644)
	if err != nil {
		return err
	}

	return nil
}

func convert(mOld *mv1.Manifest) (*manifest.Manifest, error) {
	var services manifest.Services

	for name, service := range mOld.Services {
		// build
		b := manifest.ServiceBuild{
			Path: service.Build.Context,
		}

		// build args
		if len(service.Build.Args) > 0 {
			fmt.Println("WARNING: Build args are not supported in convox.yml. Use ARG in your Dockerfile instead.")
		}

		// build dockerfile
		if service.Build.Dockerfile != "" {
			fmt.Println("WARNING: The dockerfile key is not supported in convox.yml. Please rename your file to \"Dockerfile\".")
		}

		// command
		var cmd manifest.ServiceCommand
		if len(service.Command.Array) > 0 {
			cmd.Development = strings.Join(service.Command.Array, " ")
			cmd.Production = strings.Join(service.Command.Array, " ")
		} else {
			cmd.Development = service.Command.String
			cmd.Production = service.Command.String
		}

		// cpu_shares
		if service.Cpu != 0 {
			fmt.Println("INFO: cpu_shares are not configurable via convox.yml")
		}

		// entrypoint
		if service.Entrypoint != "" {
			fmt.Println("WARNING: The entrypoint key is not supported in convox.yml. Use ENTRYPOINT in your Dockerfile instead.")
		}

		// environment
		env := []string{}
		for _, eItem := range service.Environment {
			if eItem.Needed {
				env = append(env, eItem.Name)
			} else {
				env = append(env, fmt.Sprintf("%s=%s", eItem.Name, eItem.Value))
			}
		}

		s := manifest.Service{
			Name:        name,
			Build:       b,
			Command:     cmd,
			Environment: env,
		}
		services = append(services, s)
	}

	m := manifest.Manifest{
		Services: services,
	}

	return &m, nil
}
