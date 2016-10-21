package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/convox/praxis/cli"
	"github.com/convox/praxis/client"
	"github.com/docker/docker/pkg/archive"
)

func init() {
	cli.Register(cli.Command{
		Func:    cmdBuild,
		Name:    "build",
		Usage:   "[source]",
		Summary: "build an image",
		Description: `
			lorem ipsum dolor
			  sit amet

			consectetur adipiscing elit
		`,
	})
}

func cmdBuild(c cli.Context) error {
	build, err := buildDirectory("test", ".")
	if err != nil {
		return err
	}

	fmt.Printf("build = %+v\n", build)

	return nil
}

func buildDirectory(app, dir string) (*client.Build, error) {
	url, err := uploadDirectory(app, ".")
	if err != nil {
		return nil, err
	}

	build, err := rack().BuildCreate(app, url, client.BuildCreateOptions{})
	if err != nil {
		return nil, err
	}

	id := build.Id

	r, err := rack().BuildLogs(app, build.Id)
	fmt.Printf("r = %+v\n", r)
	fmt.Printf("err = %+v\n", err)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(os.Stdout, r); err != nil {
		return nil, err
	}

	for {
		build, err := rack().BuildGet(app, id)
		if err != nil {
			return nil, err
		}

		switch build.Status {
		case "complete":
			return build, nil
		case "error":
			return nil, fmt.Errorf("build failed")
		}

		time.Sleep(1 * time.Second)
	}
}

func createTarball(dir string) (io.Reader, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	defer os.Chdir(cwd)

	sym, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, err
	}

	if err := os.Chdir(sym); err != nil {
		return nil, err
	}

	options := &archive.TarOptions{
		Compression: archive.Gzip,
	}

	return archive.TarWithOptions(sym, options)
}

func uploadDirectory(app, dir string) (string, error) {
	r, err := createTarball(dir)
	if err != nil {
		return "", err
	}

	return rack().BlobStore(app, "", r, client.BlobStoreOptions{
		Public: false,
	})
}
