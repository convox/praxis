package local

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/convox/logger"
)

var (
	customTopic       = os.Getenv("CUSTOM_TOPIC")
	notificationTopic = os.Getenv("NOTIFICATION_TOPIC")
	sortableTime      = "20060102.150405.000000000"
)

// Logger is a package-wide logger
var Logger = logger.New("ns=provider.aws")

type Provider struct {
	Root string
}

// FromEnv returns a new local.Provider from env vars
func FromEnv() (*Provider, error) {
	return &Provider{Root: coalesce(os.Getenv("PROVIDER_ROOT"), "/var/convox")}, nil
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func (p *Provider) Delete(key string) error {
	if p.Root == "" {
		return fmt.Errorf("cannot delete with empty root")
	}

	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("no such key: %s", key)
	}

	return os.Remove(path)
}

func (p *Provider) DeleteAll(key string) error {
	if p.Root == "" {
		return fmt.Errorf("cannot delete with empty root")
	}

	return os.RemoveAll(filepath.Join(p.Root, key))
}

func (p *Provider) Exists(key string) bool {
	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func (p *Provider) Read(key string) (io.ReadCloser, error) {
	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("no such key: %s", key)
	}

	return os.Open(path)
}

func (p *Provider) List(key string) ([]string, error) {
	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return nil, err
	}

	fd, err := os.Open(path)
	if err != nil {
		return []string{}, nil
	}

	return fd.Readdirnames(-1)
}

func (p *Provider) Load(key string, v interface{}) error {
	r, err := p.Read(key)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (p *Provider) Store(key string, v interface{}) error {
	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	if r, ok := v.(io.Reader); ok {
		fd, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}

		defer fd.Close()

		if _, err := io.Copy(fd, r); err != nil {
			return err
		}

		return nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0600)
}

func (p *Provider) Logs(pid string) (io.ReadCloser, error) {
	r, w := io.Pipe()

	cmd := exec.Command("docker", "logs", "--follow", pid)

	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		cmd.Wait()
		w.Close()
	}()

	return r, nil
}

func (p *Provider) write(key string) (io.WriteCloser, error) {
	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}

	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_SYNC|os.O_APPEND, 0600)
}

func (p *Provider) tail(key string) (io.ReadCloser, error) {
	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	cmd := exec.Command("tail", "-f", path)
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return r, nil
}
