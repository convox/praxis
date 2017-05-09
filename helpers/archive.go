package helpers

import (
	"io"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"
)

type TarballOptions struct {
	Includes []string
	Excludes []string
}

func CreateTarball(dir string, opts TarballOptions) (io.ReadCloser, error) {
	sym, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, err
	}

	if len(opts.Includes) == 0 {
		opts.Includes = []string{"."}
	}

	topts := &archive.TarOptions{
		Compression:     archive.Gzip,
		ExcludePatterns: opts.Excludes,
		IncludeFiles:    opts.Includes,
	}

	return archive.TarWithOptions(sym, topts)
}

func ExtractTarball(r io.Reader, dir string) error {
	cmd := exec.Command("tar", "xzvf", "-", "-C", dir)

	cmd.Stdin = r
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	return cmd.Run()
}
