package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "test",
		Description: "run tests",
		Action:      runTest,
	})
}

func runTest(c *cli.Context) error {
	name := fmt.Sprintf("test-%d", time.Now().Unix())

	app, err := Rack.AppCreate(name)
	if err != nil {
		return err
	}

	defer Rack.AppDelete(name)

	release, err := releaseDirectory(app.Name, ".")
	if err != nil {
		return err
	}

	fmt.Printf("release = %+v\n", release)

	data, err := Rack.ReleaseManifest(app.Name, release.Id)
	if err != nil {
		return err
	}

	fmt.Printf("len(data) = %+v\n", len(data))

	m, err := manifest.Load(data)
	if err != nil {
		return err
	}

	fmt.Printf("m = %+v\n", m)

	for _, s := range m.Services {
		if s.Test != "" {
			_, err := Rack.ProcessRun(app.Name, s.Name, rack.ProcessRunOptions{
				Output: os.Stdout,
			})
			if err != nil {
				return err
			}
		}
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
func releaseDirectory(app, dir string) (*rack.Release, error) {
	r, err := createTarball(dir)
	if err != nil {
		return nil, err
	}

	object, err := Rack.ObjectStore(app, "", r)
	if err != nil {
		return nil, err
	}

	build, err := Rack.BuildCreate(app, fmt.Sprintf("object://%s", object.Key))
	if err != nil {
		return nil, err
	}

	return Rack.ReleaseGet(app, build.Release)
}
