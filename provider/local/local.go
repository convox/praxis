package local

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

var (
	Docker   *docker.Client
	SystemId string
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	dc, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		panic(err)
	}

	Docker = dc

	SystemId = fmt.Sprintf("%d", rand.Int63())
}

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

func (p *Provider) remove(key string) error {
	return os.Remove(filepath.Join(p.Home, key))
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
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return prefix + string(b)
}
