package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/convox/praxis/changes"
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
	if err := startLocalRack(); err != nil {
		return err
	}

	name := "test"

	app, err := Rack.AppGet(name)
	if err != nil {
		return err
	}

	m, err := manifest.LoadFile("convox.yml")
	if err != nil {
		return err
	}

	ch := make(chan error)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go handleSignals(sig, ch, m, app.Name)

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
		w := m.PrefixWriter(os.Stdout, s.Name)
		go startService(app.Name, s.Name, build.Release, w, ch)
		go watchChanges(m, s.Name, ch)
	}

	for _, b := range m.Balancers {
		go startBalancer(app.Name, b, ch)
	}

	return <-ch
}

func handleSignals(ch chan os.Signal, errch chan error, m *manifest.Manifest, app string) {
	sig := <-ch

	if sig == syscall.SIGINT {
		fmt.Println("")
	}

	w := m.PrefixWriter(os.Stdout, "convox")

	w.Writef("stopping\n")

	ps, err := Rack.ProcessList(app, types.ProcessListOptions{})
	if err != nil {
		errch <- err
		return
	}

	var wg sync.WaitGroup

	wg.Add(len(ps))

	for _, p := range ps {
		w.Writef("stopping %s.%s\n", p.Service, p.Id)
		go func() {
			defer wg.Done()
			Rack.ProcessStop(app, p.Id)
		}()
	}

	wg.Wait()

	w.Writef("stopped\n")

	os.Exit(0)
}

func startService(app, service, release string, w manifest.PrefixWriter, ch chan error) {
	w.Writef("starting\n")

	_, err := Rack.ProcessRun(app, types.ProcessRunOptions{
		Release: release,
		Service: service,
		Stream: types.Stream{
			Reader: nil,
			Writer: w,
		},
	})

	ch <- err
}

func startBalancer(app string, balancer manifest.Balancer, ch chan error) {
	for _, e := range balancer.Endpoints {
		name := fmt.Sprintf("balancer-%s-%s-%s", app, balancer.Name, e.Port)

		exec.Command("docker", "rm", "-f", name).Run()

		args := []string{"run"}

		args = append(args, "--rm", "--name", name)
		args = append(args, "-p", fmt.Sprintf("%s:3000", e.Port))
		args = append(args, "--link", "rack")
		args = append(args, "-e", "RACK_URL=https://rack:3000")
		args = append(args, "-e", fmt.Sprintf("APP=%s", app))
		args = append(args, "convox/praxis", "proxy")
		args = append(args, e.Protocol)

		switch {
		case e.Redirect != "":
			args = append(args, "redirect", e.Redirect)
		case e.Target != "":
			args = append(args, "target", e.Target)
		default:
			ch <- fmt.Errorf("invalid balancer endpoint: %s:%s", balancer.Name, e.Port)
			return
		}

		// if p.Secure {
		//   args = append(args, "secure")
		// }

		cmd := exec.Command("docker", args...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			ch <- err
			return
		}
	}
}

func watchChanges(m *manifest.Manifest, service string, ch chan error) {
	for _, s := range m.Services {
		bss, err := m.BuildSources(s.Name)
		if err != nil {
			ch <- err
			return
		}

		for _, bs := range bss {
			go watchPath(m, s.Name, bs, ch)
		}
	}
}

func watchPath(m *manifest.Manifest, service string, bs manifest.BuildSource, ch chan error) {
	cch := make(chan changes.Change, 1)

	changes.Watch(bs.Local, cch)

	for c := range cch {
		fmt.Printf("c = %+v\n", c)
	}
}
