package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
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

	// defer Rack.AppDelete(name)

	release, err := buildDirectory(app.Name, ".")
	if err != nil {
		return err
	}

	build, err := Rack.BuildGet(app.Name, release.Build)
	if err != nil {
		return err
	}

	m, err := manifest.Load([]byte(build.Manifest))
	if err != nil {
		return err
	}

	for _, s := range m.Services {
		if s.Test != "" {
			w := m.PrefixWriter(os.Stdout, s.Name)

			if _, err := w.Write([]byte(fmt.Sprintf("running: %s\n", s.Test))); err != nil {
				return err
			}

			err := Rack.ProcessRun(app.Name, types.ProcessRunOptions{
				Command: s.Test,
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

func buildDirectory(app, dir string) (*types.Release, error) {
	r, err := createTarball(dir)
	if err != nil {
		return nil, err
	}

	object, err := Rack.ObjectStore(app, "", r)
	if err != nil {
		return nil, err
	}

	build, err := Rack.BuildCreate(app, fmt.Sprintf("object:///%s", object.Key))
	if err != nil {
		return nil, err
	}

	m, err := manifest.Load([]byte(build.Manifest))
	if err != nil {
		return nil, err
	}

	// for {
	//   build, err := Rack.BuildGet(app, build.Id)
	//   if err != nil {
	//     return nil, err
	//   }

	//   fmt.Printf("build = %+v\n", build)

	//   break
	// }

	logs, err := Rack.BuildLogs(app, build.Id)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(types.Stream{Writer: m.PrefixWriter(os.Stdout, "build")}, logs); err != nil {
		return nil, err
	}

	build, err = Rack.BuildGet(app, build.Id)
	if err != nil {
		return nil, err
	}

	return Rack.ReleaseGet(app, build.Release)
}
