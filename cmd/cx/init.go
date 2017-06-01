package main

import (
	"io/ioutil"

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

	for name, _ := range mOld.Services {
		service := manifest.Service{
			Name: name,
		}
		services = append(services, service)
	}

	m := manifest.Manifest{
		Services: services,
	}

	return &m, nil
}
