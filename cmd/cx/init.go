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

		//TODO: labels
		//TODO: links

		// mem_limit
		mb := service.Memory / (1024 * 1024) // bytes to Megabytes
		scale := manifest.ServiceScale{
			Memory: int(mb),
		}

		// ports
		p := manifest.ServicePort{}
		if len(service.Ports) > 1 {
			fmt.Printf("WARNING: Multiple ports found for %s. Only 1 HTTP port per service is supported.\n", service.Name)
		}
		for _, port := range service.Ports {
			if port.Protocol == "udp" {
				fmt.Printf("WARNING: %s %s - UDP ports are not supported.\n", service.Name, port)
				continue
			}
			switch port.Balancer {
			case 80:
				p.Port = port.Container
				p.Scheme = "http"
			case 443:
				if p.Port != 80 {
					p.Port = port.Container
					p.Scheme = "https"
				}
			default:
				fmt.Printf("WARNING: %s %s - Only HTTP ports supported.\n", service.Name, port)
			}
		}

		s := manifest.Service{
			Name:        name,
			Build:       b,
			Command:     cmd,
			Environment: env,
			Image:       service.Image,
			Port:        p,
			Scale:       scale,
		}
		services = append(services, s)
	}

	m := manifest.Manifest{
		Services: services,
	}

	err := m.ApplyDefaults()
	if err != nil {
		return nil, err
	}

	return &m, nil
}
