package local

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/convox/logger"
	homedir "github.com/mitchellh/go-homedir"
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

// NewProviderFromEnv returns a new AWS provider from env vars
func FromEnv() *Provider {
	home, err := homedir.Expand("~/.convox/local")
	if err != nil {
		panic(err)
	}

	return &Provider{Root: home}
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
