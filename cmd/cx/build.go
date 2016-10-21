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

	return rack().BuildCreate(app, url, client.BuildCreateOptions{})
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

func streamBuild(app, build string, w io.Writer) error {
	r, err := rack().BuildLogs(app, build)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	return nil
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

func waitForBuild(app, build string) (*client.Build, error) {
	for {
		b, err := rack().BuildGet(app, build)
		if err != nil {
			return nil, err
		}

		switch b.Status {
		case "complete":
			return b, nil
		case "error":
			return nil, fmt.Errorf("build failed")
		}

		time.Sleep(1 * time.Second)
	}
}
