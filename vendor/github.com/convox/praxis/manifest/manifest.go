package manifest

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type Manifest struct {
	Balancers Balancers
	Queues    Queues
	Services  Services
	Tables    Tables
	Timers    Timers

	root string
}

func Load(data []byte) (*Manifest, error) {
	var m Manifest

	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

func LoadFile(path string) (*Manifest, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	m, err := Load(data)
	if err != nil {
		return nil, err
	}

	root, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	m.root = root

	return m, nil
}

func LoadEnvironment(file string) ([]string, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return []string{}, nil
	}

	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	env := []string{}

	s := bufio.NewScanner(fd)

	for s.Scan() {
		env = append(env, s.Text())
	}

	return env, nil
}

func (m *Manifest) Path(sub string) (string, error) {
	if m.root == "" {
		return "", fmt.Errorf("path undefined for a manifest with no root")
	}

	return filepath.Join(m.root, sub), nil
}

func (m *Manifest) Validate(env []string) error {
	for _, s := range m.Services {
		if err := s.Validate(env); err != nil {
			return err
		}
	}

	return nil
}

func message(w io.Writer, format string, args ...interface{}) {
	if w != nil {
		w.Write([]byte(fmt.Sprintf(format, args...) + "\n"))
	}
}
