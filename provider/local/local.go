package local

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
)

type Provider struct {
	Home string
}

func FromEnv() *Provider {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	return &Provider{
		Home: filepath.Join(u.HomeDir, ".convox", "praxis"),
	}
}

func (p *Provider) load(key string) (io.ReadCloser, error) {
	return nil, nil
}

func (p *Provider) save(key string, r io.Reader) error {
	file := filepath.Join(p.Home, key)
	dir := filepath.Dir(file)

	dstat, err := os.Stat(dir)
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	case err != nil:
		return err
	case os.IsExist(err) && !dstat.IsDir():
		return fmt.Errorf("file exists: %s", dir)
	}

	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if _, err := io.Copy(fd, r); err != nil {
		return err
	}

	return nil
}
