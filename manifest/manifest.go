package manifest

import (
	"fmt"
	"io"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const (
	StageProduction  = 0
	StageDevelopment = iota
	StageTest        = iota
)

type Manifest struct {
	Balancers   Balancers
	Environment Environment
	Keys        Keys
	Queues      Queues
	Resources   Resources
	Services    Services
	Tables      Tables
	Timers      Timers
	Workflows   Workflows
}

func Load(data []byte, env Environment) (*Manifest, error) {
	var m Manifest

	p, err := interpolate(data, env)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(p, &m); err != nil {
		return nil, err
	}

	m.Environment = env

	if err := m.applyDefaults(); err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Manifest) Service(name string) (*Service, error) {
	for _, s := range m.Services {
		if s.Name == name {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("no such service: %s", name)
}

func (m *Manifest) ServiceEnvironment(service string) (Environment, error) {
	s, err := m.Service(service)
	if err != nil {
		return nil, err
	}

	env := Environment{}

	for _, e := range s.Environment {
		parts := strings.SplitN(e, "=", 2)

		switch len(parts) {
		case 1:
			v, ok := m.Environment[parts[0]]
			if !ok {
				return nil, fmt.Errorf("required env: %s", parts[0])
			}
			env[parts[0]] = v
		case 2:
			v, ok := m.Environment[parts[0]]
			if ok {
				env[parts[0]] = v
			} else {
				env[parts[0]] = parts[1]
			}
		default:
			return nil, fmt.Errorf("invalid environment declaration: %s", e)
		}
	}

	return env, nil
}

func (m *Manifest) Validate() error {
	// for _, s := range m.Services {
	//   if err := s.Validate(m.Environment); err != nil {
	//     return err
	//   }
	// }

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
