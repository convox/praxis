package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/convox/praxis/changes"
	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	reAppLog = regexp.MustCompile(`^\[([^.]+)\.([^\]]+)\] (.*)$`)
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "start",
		Description: "start the app in development mode",
		Action:      runStart,
		Flags: []cli.Flag{
			appFlag,
		},
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

	ch := make(chan error)

	bw := types.Stream{Writer: os.Stdout}

	b, err := buildDirectory(app, ".", types.BuildCreateOptions{Stage: manifest.StageDevelopment}, bw)
	if err != nil {
		return err
	}

	if err := buildLogs(b, bw); err != nil {
		return err
	}

	b, err = Rack.BuildGet(app, b.Id)
	if err != nil {
		return err
	}

	switch b.Status {
	case "created", "running", "complete":
	case "failed":
		return fmt.Errorf("build failed")
	default:
		return fmt.Errorf("unknown build status: %s", b.Status)
	}

	m, _, err := helpers.AppManifest(Rack, app)
	if err != nil {
		return err
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go handleSignals(sig, ch, m, app)

	rw := types.Stream{Writer: m.Writer("release", os.Stdout)}

	if err := releaseLogs(app, b.Release, rw); err != nil {
		return err
	}

	r, err := Rack.ReleaseGet(app, b.Release)
	if err != nil {
		return err
	}

	switch r.Status {
	case "created", "current", "promoting", "promoted":
	case "failed":
		return fmt.Errorf("release failed")
	default:
		return fmt.Errorf("unknown release status: %s", r.Status)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	for _, s := range m.Services {
		go watchChanges(wd, m, app, s.Name, ch)
	}

	logs, err := Rack.AppLogs(app, types.LogsOptions{Follow: true})
	if err != nil {
		return err
	}

	ls := bufio.NewScanner(logs)

	for ls.Scan() {
		match := reAppLog.FindStringSubmatch(ls.Text())

		if len(match) == 4 {
			m.Writef(match[1], "%s\n", match[3])
		}
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

func watchChanges(root string, m *manifest.Manifest, app, service string, ch chan error) {
	for _, s := range m.Services {
		bss, err := m.BuildSources(root, s.Name)
		if err != nil {
			ch <- err
			return
		}

		for _, bs := range bss {
			go watchPath(root, m, app, s.Name, bs, ch)
		}
	}
}

func watchPath(root string, m *manifest.Manifest, app, service string, bs manifest.BuildSource, ch chan error) {
	cch := make(chan changes.Change, 1)

	w := m.Writer("convox", os.Stdout)

	abs, err := filepath.Abs(bs.Local)
	if err != nil {
		ch <- err
		return
	}

	ignores, err := m.BuildIgnores(root, service)
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
					w.Writef("sync: %d files to %s\n", len(adds), ps.Service)
				case len(adds) > 0:
					w.Writef("sync: %s to %s\n", strings.Join(changes.Files(adds), ", "), ps.Service)
				}

				if err := handleAdds(app, ps.Id, bs.Remote, adds); err != nil {
					w.Writef("sync add error: %s\n", err)
				}

				switch {
				case len(removes) > 3:
					w.Writef("remove: %d files to %s\n", len(removes), ps.Service)
				case len(removes) > 0:
					w.Writef("remove: %s to %s\n", strings.Join(changes.Files(removes), ", "), ps.Service)
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
