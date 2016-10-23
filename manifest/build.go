package manifest

import (
	"crypto/sha256"
	"fmt"

	"github.com/convox/praxis/exec"
)

type BuildOptions struct {
	Cache  bool
	Prefix string
}

func (m *Manifest) Build(opts BuildOptions) error {
	builds := map[string][]string{}
	pulls := map[string][]string{}

	for _, s := range m.Services {
		switch {
		case s.Build != "":
			builds[s.Build] = append(builds[s.Build], s.Name)
		case s.Image != "":
			pulls[s.Image] = append(pulls[s.Image], s.Name)
		}
	}

	for dir, tags := range builds {
		id := fmt.Sprintf("%x", sha256.Sum256([]byte(dir)))[0:10]

		cmd := exec.Command("docker", "build", "-t", id, dir)

		pw := m.Prefix("build")

		cmd.Stdout = pw
		cmd.Stderr = Writer

		m.system("building %s", dir)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build failed")
		}

		for _, tag := range tags {
			if err := exec.Command("docker", "tag", id, tag).Run(); err != nil {
				return fmt.Errorf("could not tag: %s", tag)
			}
		}
	}

	for image, tags := range pulls {
		if err := exec.Command("docker", "pull", image).Run(); err != nil {
			return fmt.Errorf("could not pull: %s", image)
		}

		for _, tag := range tags {
			if err := exec.Command("docker", "tag", image, tag).Run(); err != nil {
				return fmt.Errorf("could not tag: %s", tag)
			}
		}
	}

	return nil
}
