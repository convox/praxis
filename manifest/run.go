package manifest

import "os/exec"

func (m *Manifest) Run() error {
	for _, s := range m.Services {
		cmd := exec.Command("docker", "run", "-i", s.Name)

		pw := m.prefix(s.Name)

		cmd.Stdout = pw
		cmd.Stderr = pw

		if err := cmd.Run(); err != nil {
			return nil
		}
	}

	return nil
}
