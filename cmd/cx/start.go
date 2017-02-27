package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "start",
		Description: "start the app in development mode",
		Action:      runStart,
	})
}

func runStart(c *cli.Context) error {
	name := "test"

	app, err := Rack.AppGet(name)
	if err != nil {
		return err
	}

	m, err := manifest.LoadFile("convox.yml")
	if err != nil {
		return err
	}

	for _, b := range m.Balancers {
		if err := startBalancer(app.Name, b); err != nil {
			return err
		}
	}

	time.Sleep(1 * time.Second)
	os.Exit(1)

	build, err := buildDirectory(app.Name, ".")
	if err != nil {
		return err
	}

	if err := buildLogs(build, types.Stream{Writer: m.PrefixWriter(os.Stdout, "build")}); err != nil {
		return err
	}

	build, err = Rack.BuildGet(app.Name, build.Id)
	if err != nil {
		return err
	}

	for _, s := range m.Services {
		if s.Test == "" {
			continue
		}

		w := m.PrefixWriter(os.Stdout, s.Name)

		w.Writef("starting\n")

		err := Rack.ProcessRun(app.Name, types.ProcessRunOptions{
			Service: s.Name,
			Stream: types.Stream{
				Reader: nil,
				Writer: w,
			},
		})
		if err != nil {
			return err
		}
	}

	// proxy

	// changes

	return nil
}

func startBalancer(app string, balancer manifest.Balancer) error {
	fmt.Printf("balancer = %+v\n", balancer)

	for _, e := range balancer.Endpoints {
		fmt.Printf("e = %+v\n", e)

		name := fmt.Sprintf("%s-%s-%s", app, balancer.Name, e.Port)

		exec.Command("docker", "rm", "-f", name).Run()

		args := []string{"run"}

		args = append(args, "--rm", "--name", name)
		args = append(args, "-p", fmt.Sprintf("%s:3000", e.Port))
		args = append(args, "convox/praxis", "proxy")
		args = append(args, e.Protocol)

		switch {
		case e.Redirect != "":
			args = append(args, "redirect", e.Redirect)
		case e.Target != "":
			args = append(args, "target", e.Target)
		default:
			return fmt.Errorf("invalid balancer endpoint: %s:%s", balancer.Name, e.Port)
		}

		// if p.Secure {
		//   args = append(args, "secure")
		// }

		fmt.Printf("args = %+v\n", args)

		cmd := exec.Command("docker", args...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
