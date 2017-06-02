package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
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
	services := manifest.Services{}
	timers := make(manifest.Timers, 0)

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

		// convox.agent
		if service.IsAgent() {
			fmt.Printf("WARNING: %s - Running a service as an agent is not supported.\n", service.Name)
		}

		// convox.balancer
		if (len(service.Ports) > 0) && !service.HasBalancer() {
			fmt.Printf("WARNING: %s - Disabling balancers with convox.balancer=false is not supported.\n", service.Name)
		}

		// convox.cron
		for k, v := range service.LabelsByPrefix("convox.cron") {
			timer := manifest.Timer{}
			ks := strings.Split(k, ".")
			tokens := strings.Fields(v)
			timer.Name = ks[len(ks)-1]
			timer.Command = strings.Join(tokens[5:], " ")
			timer.Schedule = strings.Join(tokens[0:5], " ")
			timer.Service = service.Name
			timers = append(timers, timer)
		}

		// convox.deployment.maximum
		if len(service.LabelsByPrefix("convox.deployment.maximum")) > 0 {
			fmt.Printf("WARNING: %s - Setting deployment maximum is not supported.\n", service.Name)
		}

		// convox.deployment.minimum
		if len(service.LabelsByPrefix("convox.deployment.minimum")) > 0 {
			fmt.Printf("WARNING: %s - Setting deployment minimum is not supported.\n", service.Name)
		}

		// convox.draining.timeout
		if len(service.LabelsByPrefix("convox.draining.timeout")) > 0 {
			fmt.Printf("WARNING: %s - Setting draning timeout is not supported.\n", service.Name)
		}

		// convox.environment.secure
		if len(service.LabelsByPrefix("convox.environment.secure")) > 0 {
			fmt.Printf("INFO: %s - Declaring secure environment is not necessary. Praxis environments are secure by default.\n", service.Name)
		}

		// convox.health.path
		// convox.health.timeout
		health := manifest.ServiceHealth{}
		if balancer := mOld.GetBalancer(service.Name); balancer != nil {
			timeout, err := strconv.Atoi(balancer.HealthTimeout())
			if err != nil {
				fmt.Println("Well, shit.")
			}
			health.Path = balancer.HealthPath()
			health.Timeout = timeout
		}

		// convox.health.port
		if len(service.LabelsByPrefix("convox.health.port")) > 0 {
			fmt.Printf("INFO: %s - Setting health check port is not necessary.\n", service.Name)
		}

		// convox.health.threshold.healthy
		// convox.helath.threshold.unhealthy
		if len(service.LabelsByPrefix("convox.health.threshold")) > 0 {
			fmt.Printf("INFO: %s - Setting health check thresholds is not supported.\n", service.Name)
		}

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

		// privileged
		if service.Privileged {
			fmt.Printf("WARNING: %s - Privileged mode not supported.\n", service.Name)
		}

		s := manifest.Service{
			Name:        name,
			Build:       b,
			Command:     cmd,
			Environment: env,
			Health:      health,
			Image:       service.Image,
			Port:        p,
			Scale:       scale,
			Volumes:     service.Volumes,
		}
		services = append(services, s)
	}

	m := manifest.Manifest{
		Services: services,
		Timers:   timers,
	}

	err := m.ApplyDefaults()
	if err != nil {
		return nil, err
	}

	return &m, nil
}
