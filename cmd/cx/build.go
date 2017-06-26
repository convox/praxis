package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	"github.com/docker/docker/builder/dockerignore"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "build",
		Description: "build an application",
		Action:      runBuild,
		Flags: []cli.Flag{
			appFlag,
		},
	})
	stdcli.RegisterCommand(cli.Command{
		Name:        "builds",
		Description: "list builds",
		Action:      runBuilds,
		Flags: []cli.Flag{
			appFlag,
		},
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "logs",
				Description: "show build logs",
				Action:      runBuildsLogs,
				Flags: []cli.Flag{
					appFlag,
				},
			},
		},
	})
}

func runBuild(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	_, err = buildDirectory(app, ".", types.BuildCreateOptions{}, os.Stdout)
	if err != nil {
		return err
	}

	return nil
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

func runBuildsLogs(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return stdcli.Usage(c)
	}

	id := c.Args()[0]

	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	logs, err := Rack.BuildLogs(app, id)
	if err != nil {
		return err
	}

	if _, err := io.Copy(os.Stdout, logs); err != nil {
		return err
	}

	return nil
}

func buildDirectory(app, dir string, opts types.BuildCreateOptions, w io.Writer) (*types.Build, error) {
	if _, err := Rack.AppGet(app); err != nil {
		return nil, err
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	sw := *stdcli.DefaultWriter

	sw.Stdout = w
	sw.Stderr = w

	sw.Writef("<start>building:</start> <dir>%s</dir>\n", abs)
	sw.Startf("uploading")

	r, err := createTarball(dir)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	object, err := Rack.ObjectStore(app, "", r, types.ObjectStoreOptions{})
	if err != nil {
		return nil, err
	}

	sw.OK()

	sw.Startf("starting build")

	build, err := Rack.BuildCreate(app, fmt.Sprintf("object:///%s", object.Key), opts)
	if err != nil {
		return nil, err
	}

	if err := tickWithTimeout(2*time.Second, 5*time.Minute, notBuildStatus(app, build.Id, "created")); err != nil {
		return nil, err
	}

	build, err = Rack.BuildGet(app, build.Id)
	if err != nil {
		return nil, err
	}

	sw.Writef("<id>%s</id>\n", build.Process)

	if err := buildLogs(build, stdcli.TagWriter("log", w)); err != nil {
		return nil, err
	}

	build, err = Rack.BuildGet(app, build.Id)
	if err != nil {
		return nil, err
	}

	if build.Status == "failed" {
		return nil, fmt.Errorf("build failed")
	}

	return build, nil
}

func buildLogs(build *types.Build, w io.Writer) error {
	logs, err := Rack.BuildLogs(build.App, build.Id)
	if err != nil {
		return err
	}

	go io.Copy(w, logs)

	for {
		b, err := Rack.BuildGet(build.App, build.Id)
		if err != nil {
			return err
		}

		if b.Status != "running" {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}

func createTarball(dir string) (io.ReadCloser, error) {
	excludes := []string{}

	sym, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, err
	}

	abs, err := filepath.Abs(sym)
	if err != nil {
		return nil, err
	}

	if fd, err := os.Open(filepath.Join(abs, ".dockerignore")); err == nil {
		if e, err := dockerignore.ReadAll(fd); err != nil {
			return nil, err
		} else {
			excludes = e
		}
	}

	return helpers.CreateTarball(dir, helpers.TarballOptions{Excludes: excludes})
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
		stop, err := fn()
		if err != nil {
			return err
		}
		if stop {
			return nil
		}

		select {
		case <-tickch:
			continue
		case <-timeoutch:
			return fmt.Errorf("timeout")
		}
	}
}
