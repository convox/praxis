package manifest

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kr/text"
)

type BuildOptions struct {
	Root   string
	Stdout io.Writer
	Stderr io.Writer
}

func (m *Manifest) Build(app string, opts BuildOptions) error {
	builds := map[string]Service{}
	tags := map[string]string{}

	for _, s := range m.Services {
		hash := s.BuildHash()
		builds[hash] = s
		tags[hash] = fmt.Sprintf("%s/%s", app, s.Name)
	}

	for hash, service := range builds {
		if service.Image != "" {
			if err := service.pull(tags[hash], opts); err != nil {
				return err
			}
		} else {
			if err := service.build(hash, opts); err != nil {
				return err
			}
		}
	}

	for from, to := range tags {
		if err := opts.docker("tag", from, to); err != nil {
			return err
		}
	}

	return nil
}

func (s Service) build(tag string, opts BuildOptions) error {
	if s.Build.Path == "" {
		return fmt.Errorf("must have path to build")
	}

	args := []string{"build"}

	// for _, arg := range build.Args {
	//   fmt.Printf("arg = %+v\n", arg)
	// }

	args = append(args, "--rm=false")

	args = append(args, "-t", tag)

	path, err := filepath.Abs(filepath.Join(opts.Root, s.Build.Path))
	if err != nil {
		return err
	}

	args = append(args, path)

	message(opts.Stdout, "building: %s", s.Build.Path)

	return opts.docker(args...)
}

func (s Service) pull(tag string, opts BuildOptions) error {
	if s.Image == "" {
		return fmt.Errorf("must have image to pull")
	}

	message(opts.Stdout, "pulling: %s", s.Image)

	if err := opts.docker("pull", s.Image); err != nil {
		return err
	}

	if err := opts.docker("tag", s.Image, tag); err != nil {
		return err
	}

	return nil
}

func (o BuildOptions) docker(args ...string) error {
	message(o.Stdout, "running: docker %s", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)

	cmd.Stdout = text.NewIndentWriter(o.Stdout, []byte("  "))
	cmd.Stderr = text.NewIndentWriter(o.Stderr, []byte("  "))

	return cmd.Run()
}
