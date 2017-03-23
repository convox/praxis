package local

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var lock sync.Mutex

func (p *Provider) storageDelete(key string) error {
	lock.Lock()
	defer lock.Unlock()

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

func (p *Provider) storageDeleteAll(key string) error {
	lock.Lock()
	defer lock.Unlock()

	if p.Root == "" {
		return fmt.Errorf("cannot delete with empty root")
	}

	return os.RemoveAll(filepath.Join(p.Root, key))
}

func (p *Provider) Exists(key string) bool {
	lock.Lock()
	defer lock.Unlock()

	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func (p *Provider) storageList(key string) ([]string, error) {
	lock.Lock()
	defer lock.Unlock()

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

func (p *Provider) storageLoad(key string, v interface{}) error {
	r, err := p.storageRead(key)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (p *Provider) storageRead(key string) (io.ReadCloser, error) {
	lock.Lock()
	defer lock.Unlock()

	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("no such key: %s", key)
	}

	return os.Open(path)
}

func (p *Provider) storageStore(key string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()

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

func (p *Provider) storageTail(key string) (io.ReadCloser, error) {
	lock.Lock()
	defer lock.Unlock()

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

func (p *Provider) storageWrite(key string) (io.WriteCloser, error) {
	lock.Lock()
	defer lock.Unlock()

	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}

	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_SYNC|os.O_APPEND, 0600)
}
