package local

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
)

type Provider struct {
	Home string
}

func FromEnv() *Provider {
	return &Provider{
		Home: "/var/run/convox",
	}
}

func (p *Provider) load(key string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(p.Home, key))
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

	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	if _, err := io.Copy(fd, r); err != nil {
		return err
	}

	return nil
}

const (
	idLength = 10
)

var (
	alphabet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func id(prefix string) string {
	b := make([]rune, 10)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			panic(err)
		}
		b[i] = alphabet[idx.Int64()]
	}
	return prefix + string(b)
}
