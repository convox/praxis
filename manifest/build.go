package manifest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type BuildOptions struct {
	Root   string
	Stdout io.Writer
	Stderr io.Writer
}

type BuildSource struct {
	Local  string
	Remote string
}

func (m *Manifest) Build(app string, id string, opts BuildOptions) error {
	builds := map[string]Service{}
	tags := map[string]string{}

	for _, s := range m.Services {
		hash := s.BuildHash()
		builds[hash] = s
		tags[hash] = fmt.Sprintf("%s/%s:%s", app, s.Name, id)
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

func (m *Manifest) BuildManifest(service string) ([]byte, error) {
	s, err := m.Services.Find(service)
	if err != nil {
		return nil, err
	}

	path, err := filepath.Abs(filepath.Join(m.Root, s.Build.Path, "Dockerfile"))
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(path)
}

func (m *Manifest) BuildSources(service string) ([]BuildSource, error) {
	data, err := m.BuildManifest(service)
	if err != nil {
		return nil, err
	}

	bs := []BuildSource{}
	env := map[string]string{}

	s := bufio.NewScanner(bytes.NewReader(data))

	for s.Scan() {
		parts := strings.Fields(s.Text())

		if len(parts) < 1 {
			continue
		}

		switch parts[0] {
		case "ENV":
			if len(parts) > 2 {
				env[parts[1]] = parts[2]
			}
		case "ADD", "COPY":
			if len(parts) > 2 {
				u, err := url.Parse(parts[1])
				if err != nil {
					return nil, err
				}

				switch u.Scheme {
				case "http", "https":
					// do nothing
				default:
					path := parts[2]
					for k, v := range env {
						path = strings.Replace(path, fmt.Sprintf("$%s", k), v, -1)
						path = strings.Replace(path, fmt.Sprintf("${%s}", k), v, -1)
					}
					bs = append(bs, BuildSource{Local: parts[1], Remote: path})
				}
			}
		}
	}

	return bs, nil
}

func (s Service) build(tag string, opts BuildOptions) error {
	if s.Build.Path == "" {
		return fmt.Errorf("must have path to build")
	}

	args := []string{"build"}

	// for _, arg := range build.Args {
	//   fmt.Printf("arg = %+v\n", arg)
	// }

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

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// cmd.Stdout = text.NewIndentWriter(o.Stdout, []byte("  "))
	// cmd.Stderr = text.NewIndentWriter(o.Stderr, []byte("  "))

	return cmd.Run()
}
