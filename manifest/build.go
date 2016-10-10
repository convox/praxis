package manifest

import "github.com/convox/praxis/exec"

func (m *Manifest) Build() error {
	for _, s := range m.Services {
		cmd := exec.Command("docker", "build", "-t", s.Name, s.Build)

		pw := m.prefix("build")

		cmd.Stdout = pw
		cmd.Stderr = Writer

		if err := cmd.Run(); err != nil {
			return nil
		}
	}

	return nil
}
