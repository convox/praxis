package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

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
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	if _, err := Rack.AppGet(app); err != nil {
		return err
	}

	m, err := manifest.LoadFile("convox.yml")
	if err != nil {
		return err
	}

	env, err := manifest.LoadEnvironment(".env")
	if err != nil {
		return err
	}

	if err := m.Validate(env); err != nil {
		return err
	}

	ch := make(chan error)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go handleSignals(sig, ch, m, app)

	build, err := buildDirectory(app, ".")
	if err != nil {
		return err
	}

	if err := buildLogs(build, types.Stream{Writer: m.Writer("build", os.Stdout)}); err != nil {
		return err
	}

	build, err = Rack.BuildGet(app, build.Id)
	if err != nil {
		return err
	}

	if err := Rack.ReleasePromote(app, build.Release); err != nil {
		return err
	}

	switch build.Status {
	case "created", "running", "complete":
	case "failed":
		return fmt.Errorf("build failed")
	default:
		return fmt.Errorf("unknown build status: %s", build.Status)
	}

	for _, s := range m.Services {
		m.Writef("convox", "starting: %s\n", s.Name)

		go startService(m, app, s.Name, build.Release, env, ch)
		go watchChanges(m, app, s.Name, ch)
	}

	for _, b := range m.Balancers {
		go startBalancer(app, b, ch)
	}

	return <-ch
}

func handleSignals(ch chan os.Signal, errch chan error, m *manifest.Manifest, app string) {
	sig := <-ch

	if sig == syscall.SIGINT {
		fmt.Println("")
	}

	ps, err := Rack.ProcessList(app, types.ProcessListOptions{})
	if err != nil {
		errch <- err
		return
	}

	var wg sync.WaitGroup

	wg.Add(len(ps))

	for _, p := range ps {
		m.Writef("convox", "stopping %s\n", p.Id)

		go func() {
			defer wg.Done()
			Rack.ProcessStop(app, p.Id)
		}()
	}

	wg.Wait()

	m.Writef("convox", "stopped\n")

	os.Exit(0)
}

func startService(m *manifest.Manifest, app, service, release string, env []string, ch chan error) {
	w := m.Writer(service, os.Stdout)

	pss, err := Rack.ProcessList(app, types.ProcessListOptions{Service: service})
	if err != nil {
		ch <- err
		return
	}

	for _, ps := range pss {
		if err := Rack.ProcessStop(app, ps.Id); err != nil {
			ch <- err
			return
		}
	}

	s, err := m.Services.Find(service)
	if err != nil {
		ch <- err
		return
	}

	senv, err := s.Env(env)
	if err != nil {
		ch <- err
		return
	}

	_, err = Rack.ProcessRun(app, types.ProcessRunOptions{
		Environment: senv,
		Release:     release,
		Service:     service,
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

func watchChanges(m *manifest.Manifest, app, service string, ch chan error) {
	for _, s := range m.Services {
		bss, err := m.BuildSources(s.Name)
		if err != nil {
			ch <- err
			return
		}

		for _, bs := range bss {
			go watchPath(m, app, s.Name, bs, ch)
		}
	}
}

func watchPath(m *manifest.Manifest, app, service string, bs manifest.BuildSource, ch chan error) {
	cch := make(chan changes.Change, 1)

	w := m.Writer("convox", os.Stdout)

	abs, err := filepath.Abs(bs.Local)
	if err != nil {
		ch <- err
		return
	}

	ignores, err := m.BuildIgnores(service)
	if err != nil {
		ch <- err
		return
	}

	w.Writef("syncing: %s to %s:%s\n", bs.Local, service, bs.Remote)

	go changes.Watch(abs, cch, changes.WatchOptions{
		Ignores: ignores,
	})

	tick := time.Tick(1000 * time.Millisecond)
	chgs := []changes.Change{}

	for {
		select {
		case c := <-cch:
			chgs = append(chgs, c)
		case <-tick:
			if len(chgs) == 0 {
				continue
			}

			pss, err := Rack.ProcessList(app, types.ProcessListOptions{Service: service})
			if err != nil {
				w.Writef("sync error: %s\n", err)
				continue
			}

			adds, removes := changes.Partition(chgs)

			for _, ps := range pss {
				switch {
				case len(adds) > 3:
					w.Writef("sync: %d files\n", len(adds))
				case len(adds) > 0:
					w.Writef("sync: %s\n", strings.Join(changes.Files(adds), ", "))
				}

				if err := handleAdds(app, ps.Id, bs.Remote, adds); err != nil {
					w.Writef("sync add error: %s\n", err)
				}

				switch {
				case len(removes) > 3:
					w.Writef("remove: %d files\n", len(removes))
				case len(removes) > 0:
					w.Writef("remove: %s\n", strings.Join(changes.Files(removes), ", "))
				}

				if len(removes) > 0 {
					if err := handleRemoves(app, ps.Id, removes); err != nil {
						w.Writef("sync remove error: %s\n", err)
					}
				}
			}

			chgs = []changes.Change{}
		}
	}
}

func handleAdds(app, pid, remote string, adds []changes.Change) error {
	if len(adds) == 0 {
		return nil
	}

	r, w := io.Pipe()

	ch := make(chan error)

	go func() {
		ch <- Rack.FilesUpload(app, pid, r)
	}()

	tgz := gzip.NewWriter(w)
	tw := tar.NewWriter(tgz)

	for _, add := range adds {
		local := filepath.Join(add.Base, add.Path)

		stat, err := os.Stat(local)
		if err != nil {
			return err
		}

		tw.WriteHeader(&tar.Header{
			Name:    filepath.Join(remote, add.Path),
			Mode:    int64(stat.Mode()),
			Size:    stat.Size(),
			ModTime: stat.ModTime(),
		})

		fd, err := os.Open(local)
		if err != nil {
			return err
		}

		defer fd.Close()

		if _, err := io.Copy(tw, fd); err != nil {
			return err
		}

		fd.Close()
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if err := tgz.Close(); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return <-ch
}

func handleRemoves(app, pid string, removes []changes.Change) error {
	if len(removes) == 0 {
		return nil
	}

	return Rack.FilesDelete(app, pid, changes.Files(removes))
}
