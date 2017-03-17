package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	"github.com/convox/rack/cmd/convox/helpers"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "builds",
		Description: "list builds",
		Action:      runBuilds,
	})
	stdcli.RegisterCommand(cli.Command{
		Name:        "build",
		Description: "build the application",
		Action:      runBuild,
	})
}

func runBuilds(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	builds, err := Rack.BuildList(app)
	if err != nil {
		return err
	}

	t := stdcli.NewTable("ID", "STATUS", "STARTED", "ELAPSED")

	for _, b := range builds {
		started := helpers.HumanizeTime(b.Started)
		elapsed := stdcli.Duration(b.Started, b.Ended)

		if b.Ended.IsZero() {
			switch b.Status {
			case "running":
				elapsed = stdcli.Duration(b.Started, time.Now())
			default:
				elapsed = ""
			}
		}

		t.AddRow(b.Id, b.Status, started, elapsed)
	}

	t.Print()

	return nil
}

func runBuild(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	a, err := Rack.AppGet(app)
	if err != nil {
		return err
	}

	if a.Status != "running" {
		return fmt.Errorf("cannot build while app is %s", a.Status)
	}

	build, err := buildDirectory(app, ".")
	if err != nil {
		return err
	}

	if err := buildLogs(build, os.Stdout); err != nil {
		return err
	}

	build, err = Rack.BuildGet(app, build.Id)
	if err != nil {
		return err
	}

	return nil
}

func buildDirectory(app, dir string) (*types.Build, error) {
	if _, err := Rack.AppGet(app); err != nil {
		return nil, err
	}

	r, err := createTarball(dir)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	object, err := Rack.ObjectStore(app, "", r, types.ObjectStoreOptions{})
	if err != nil {
		return nil, err
	}

	build, err := Rack.BuildCreate(app, fmt.Sprintf("object:///%s", object.Key), types.BuildCreateOptions{})
	if err != nil {
		return nil, err
	}

	if err := tickWithTimeout(2*time.Second, 5*time.Minute, notBuildStatus(app, build.Id, "created")); err != nil {
		return nil, err
	}

	return build, nil
}

func notBuildStatus(app, id, status string) func() (bool, error) {
	return func() (bool, error) {
		build, err := Rack.BuildGet(app, id)
		if err != nil {
			return true, err
		}
		if build.Status != status {
			return true, nil
		}
		return false, nil
	}
}

func tickWithTimeout(tick time.Duration, timeout time.Duration, fn func() (stop bool, err error)) error {
	tickch := time.Tick(tick)
	timeoutch := time.After(timeout)

	for {
		select {
		case <-tickch:
			stop, err := fn()
			if err != nil {
				return err
			}
			if stop {
				return nil
			}
		case <-timeoutch:
			return fmt.Errorf("timeout")
		}
	}
}

func buildLogs(build *types.Build, w io.Writer) error {
	logs, err := Rack.BuildLogs(build.App, build.Id)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, logs); err != nil {
		return err
	}

	return nil
}

func createTarball(base string) (io.ReadCloser, error) {
	sym, err := filepath.EvalSymlinks(base)
	if err != nil {
		return nil, err
	}

	abs, err := filepath.Abs(sym)
	if err != nil {
		return nil, err
	}

	includes := []string{"."}
	excludes := []string{}

	if fd, err := os.Open(filepath.Join(abs, ".dockerignore")); err == nil {
		e, err := dockerignore.ReadAll(fd)
		if err != nil {
			return nil, err
		}

		excludes = e
	}

	options := &archive.TarOptions{
		Compression:     archive.Gzip,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	}

	return archive.TarWithOptions(sym, options)
}
