package manifest

import (
	"fmt"

	"github.com/convox/praxis/exec"
)

func (m *Manifest) Push(prefix, build string) error {
	for _, s := range m.Services {
		local := s.Name

		remote := fmt.Sprintf("%s%s:%s", prefix, s.Name, build)

		if err := exec.Command("docker", "tag", local, remote).Run(); err != nil {
			return err
		}

		if err := exec.Command("docker", "push", remote).Run(); err != nil {
			return err
		}
	}

	return nil
}
