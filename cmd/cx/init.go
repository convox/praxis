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

func resourceService(service mv1.Service) bool {
	resourceImages := []string{
		"convox/postgres",
		"convox/redis",
	}

	for _, image := range resourceImages {
		if service.Image == image {
			return true
		}
	}

	return false
}

func convert(mOld *mv1.Manifest) (*manifest.Manifest, error) {
	services := manifest.Services{}
	resources := make(manifest.Resources, 0)
	timers := make(manifest.Timers, 0)

	for name, service := range mOld.Services {
		// resources
		serviceResources := []string{}
		if resourceService(service) {
			t := ""
			switch service.Image {
			case "convox/postgres":
				t = "postgres"
			case "convox/redis":
				t = "redis"
			default:
				return nil, fmt.Errorf("%s is not a recognized resource image", service.Image)
			}
			r := manifest.Resource{
				Name: service.Name,
				Type: t,
			}
			resources = append(resources, r)
			fmt.Printf("INFO: service \"%s\" has been migrated to a resource\n", service.Name)
			continue
		}

		// build
		b := manifest.ServiceBuild{
			Path: service.Build.Context,
		}

		// build args
		if len(service.Build.Args) > 0 {
			fmt.Printf("WARN: service \"%s\" build args not migrated to convox.yml, use ARG in your Dockerfile instead\n", service.Name)
		}

		// build dockerfile
		if service.Build.Dockerfile != "" {
			fmt.Printf("WARN: service \"%s\" \"dockerfile\" key is not supported in convox.yml, file must be named \"Dockerfile\"\n", service.Name)
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
			fmt.Printf("INFO: service \"%s\" cpu_shares not configurable via convox.yml\n", service.Name)
		}

		// entrypoint
		if service.Entrypoint != "" {
			fmt.Printf("WARN: service \"%s\" \"entrypoint\" key not supported in convox.yml, use ENTRYPOINT in Dockerfile instead\n", service.Name)
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
			fmt.Printf("INFO: service \"%s\", running as an agent is not supported\n", service.Name)
		}

		// convox.balancer
		if (len(service.Ports) > 0) && !service.HasBalancer() {
			fmt.Printf("INFO: service \"%s\", disabling balancers is not supported\n", service.Name)
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
			fmt.Printf("INFO: service \"%s\", setting deployment maximum is not supported\n", service.Name)
		}

		// convox.deployment.minimum
		if len(service.LabelsByPrefix("convox.deployment.minimum")) > 0 {
			fmt.Printf("INFO: service \"%s\", setting deployment minimum is not supported\n", service.Name)
		}

		// convox.draining.timeout
		if len(service.LabelsByPrefix("convox.draining.timeout")) > 0 {
			fmt.Printf("INFO: service \"%s\", setting draning timeout is not supported\n", service.Name)
		}

		// convox.environment.secure
		if len(service.LabelsByPrefix("convox.environment.secure")) > 0 {
			fmt.Printf("INFO: service \"%s\", setting secure environment is not necessary\n", service.Name)
		}

		// convox.health.path
		// convox.health.timeout
		health := manifest.ServiceHealth{}
		if balancer := mOld.GetBalancer(service.Name); balancer != nil {
			timeout, err := strconv.Atoi(balancer.HealthTimeout())
			if err != nil {
				return nil, err
			}
			health.Path = balancer.HealthPath()
			health.Timeout = timeout
		}

		// convox.health.port
		if len(service.LabelsByPrefix("convox.health.port")) > 0 {
			fmt.Printf("INFO: service \"%s\", setting health check port is not necessary\n", service.Name)
		}

		// convox.health.threshold.healthy
		// convox.helath.threshold.unhealthy
		if len(service.LabelsByPrefix("convox.health.threshold")) > 0 {
			fmt.Printf("INFO: service \"%s\", setting health check thresholds is not supported\n", service.Name)
		}

		// convox.idle.timeout
		if len(service.LabelsByPrefix("convox.idle.timeout")) > 0 {
			fmt.Printf("INFO: service \"%s\", setting idle timeout is not supported\n", service.Name)
		}

		// convox.port..protocol
		// convox.port..proxy
		// convox.port..secure
		if len(service.LabelsByPrefix("convox.idle.timeout")) > 0 {
			fmt.Printf("INFO: service \"%s\", configuring balancer via convox.port labels is not supported\n", service.Name)
		}

		// convox.start.shift
		if len(service.LabelsByPrefix("convox.start.shift")) > 0 {
			fmt.Printf("WARN: service \"%s\", port shifting is not supported, use internal hostnames instead\n", service.Name)
		}

		// links
		for _, link := range service.Links {
			resource := false
			for _, sOld := range mOld.Services {
				if (sOld.Name == link) && resourceService(sOld) {
					serviceResources = append(serviceResources, link)
					resource = true
					break
				}
			}
			if !resource {
				fmt.Printf("WARN: service \"%s\", environment variables not generated for linked service \"%s\", use internal URL https://%s.<app name>.convox instead\n", service.Name, link, link)
			}
		}

		// mem_limit
		mb := service.Memory / (1024 * 1024) // bytes to Megabytes
		scale := manifest.ServiceScale{
			Memory: int(mb),
		}

		// ports
		p := manifest.ServicePort{}
		if len(service.Ports) > 1 {
			fmt.Printf("WARN: service \"%s\", multiple ports found, only 1 HTTP port per service is supported\n", service.Name)
		}
		for _, port := range service.Ports {
			if port.Protocol == "udp" {
				fmt.Printf("WARN: service \"%s\", UDP ports are not supported\n", service.Name)
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
				fmt.Printf("WARN: service \"%s\", only HTTP ports supported\n", service.Name)
			}
		}

		// privileged
		if service.Privileged {
			fmt.Printf("WARN: service \"%s\", privileged mode not supported\n", service.Name)
		}

		s := manifest.Service{
			Name:        name,
			Build:       b,
			Command:     cmd,
			Environment: env,
			Health:      health,
			Image:       service.Image,
			Port:        p,
			Resources:   serviceResources,
			Scale:       scale,
			Volumes:     service.Volumes,
		}
		services = append(services, s)
	}

	if mOld.Networks != nil {
		fmt.Println("INFO: custom networks not supported, use service hostnames instead")
	}

	m := manifest.Manifest{
		Resources: resources,
		Services:  services,
		Timers:    timers,
	}

	err := m.ApplyDefaults()
	if err != nil {
		return nil, err
	}

	return &m, nil
}
