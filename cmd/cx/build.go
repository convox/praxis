package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	build, err := buildDirectory(".")
	if err != nil {
		return err
	}

	fmt.Printf("build = %+v\n", build)

	return nil
}

func buildDirectory(dir string) (*client.Build, error) {
	url, err := uploadDirectory(".")
	if err != nil {
		return nil, err
	}

	build, err := rack().BuildCreate(url, client.BuildCreateOptions{})
	if err != nil {
		return nil, err
	}

	fmt.Printf("build = %+v\n", build)

	return nil, nil
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

func uploadDirectory(dir string) (string, error) {
	r, err := createTarball(dir)
	if err != nil {
		return "", err
	}

	return rack().BlobStore("", r, client.BlobStoreOptions{
		Public: false,
	})
}
