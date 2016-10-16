package manifest

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Service struct {
	Name string `yaml:"-"`

	Build       string      `yaml:"build,omitempty"`
	Command     string      `yaml:"command,omitempty"`
	Environment Environment `yaml:"environment,omitempty"`
	Image       string      `yaml:"image,omitempty"`
	Volumes     Volumes     `yaml:"volumes,omitempty"`
}

type Services []Service

func (s *Service) SyncPaths() (map[string]string, error) {
	sp := map[string]string{}

	if s.Build == "" {
		return sp, nil
	}

	data, err := ioutil.ReadFile(filepath.Join(s.Build, "Dockerfile"))
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		if len(parts) < 1 {
			continue
		}

		switch parts[0] {
		case "ADD", "COPY":
			if len(parts) >= 3 {
				sp[filepath.Join(s.Build, parts[1])] = parts[2]
			}
		}
	}

	return sp, nil
}
