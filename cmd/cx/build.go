package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "build",
		Description: "build the application",
		Action:      runBuild,
	})
}

func runBuild(c *cli.Context) error {
	app := "test"

	// a, err := Rack.AppGet(app)
	// if err != nil {
	//   return err
	// }

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
	r, err := createTarball(dir)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	object, err := Rack.ObjectStore(app, "", r)
	if err != nil {
		return nil, err
	}

	return Rack.BuildCreate(app, fmt.Sprintf("object:///%s", object.Key))
}

func buildLogs(build *types.Build, w io.Writer) error {
	// for {
	//   build, err := Rack.BuildGet(app, build.Id)
	//   if err != nil {
	//     return nil, err
	//   }

	//   fmt.Printf("build = %+v\n", build)

	//   break
	// }

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
