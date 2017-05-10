package manifest

import (
	"fmt"
	"io"

	"github.com/convox/praxis/types"

	yaml "gopkg.in/yaml.v2"
)

const (
	StageProduction  = 0
	StageDevelopment = iota
	StageTest        = iota
)

type Manifest struct {
	Balancers Balancers
	Keys      Keys
	Queues    Queues
	Resources Resources
	Services  Services
	Tables    Tables
	Timers    Timers
	Workflows Workflows
}

func Load(data []byte, env types.Environment) (*Manifest, error) {
	var m Manifest

	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	if err := m.applyDefaults(); err != nil {
		return nil, err
	}

	return &m, nil
}

// func LoadFile(path string) (*Manifest, error) {
//   data, err := ioutil.ReadFile(path)
//   if err != nil {
//     return nil, err
//   }

//   m, err := Load(data)
//   if err != nil {
//     return nil, err
//   }

//   root, err := filepath.Abs(filepath.Dir(path))
//   if err != nil {
//     return nil, err
//   }

//   m.root = root

//   return m, nil
// }

// func (m *Manifest) Path(sub string) (string, error) {
//   if m.root == "" {
//     return "", fmt.Errorf("path undefined for a manifest with no root")
//   }

//   return filepath.Join(m.root, sub), nil
// }

func (m *Manifest) Validate(env types.Environment) error {
	for _, s := range m.Services {
		if err := s.Validate(env); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manifest) applyDefaults() error {
	for i, s := range m.Services {
		if s.Build.Path == "" && s.Image == "" {
			m.Services[i].Build.Path = "."
		}

		if s.Health.Path == "" {
			m.Services[i].Health.Path = "/"
		}

		if s.Health.Interval == 0 {
			m.Services[i].Health.Interval = 5
		}

		if s.Health.Timeout == 0 {
			m.Services[i].Health.Timeout = m.Services[i].Health.Interval - 1
		}

		if s.Scale.Count.Min == 0 {
			m.Services[i].Scale.Count.Min = 1
		}

		if s.Scale.Memory == 0 {
			m.Services[i].Scale.Memory = 256
		}
	}

	// target should be inhereted for deploy and run if not set explictly
	for i, w := range m.Workflows {
		for j, s := range w.Steps {
			switch s.Type {
			case "deploy", "run":
				if s.Target == "" {
					for k := j - 1; k >= 0; k-- {
						if w.Steps[k].Target != "" {
							m.Workflows[i].Steps[j].Target = w.Steps[k].Target
						}
					}
				}
			}
		}
	}

	return nil
}

func message(w io.Writer, format string, args ...interface{}) {
	if w != nil {
		w.Write([]byte(fmt.Sprintf(format, args...) + "\n"))
	}
}
