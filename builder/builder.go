package builder

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/convox/praxis/manifest"
	docker "github.com/fsouza/go-dockerclient"
)

type Options struct {
	DockerHost string
	Namespace  string
	Stdout     io.Writer
	Stderr     io.Writer
}

type Builder struct {
	Manifest *manifest.Manifest
	Options  Options

	dockerClient *docker.Client
}

func New(m *manifest.Manifest, opts *Options) (*Builder, error) {
	b := &Builder{Manifest: m}

	if opts != nil {
		b.Options = *opts
	}

	var err error

	b.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	if opts.DockerHost != "" {
		b.dockerClient, err = docker.NewClient(b.Options.DockerHost)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (b *Builder) Build() error {
	builds := map[string]manifest.ServiceBuild{}
	pulls := map[string]bool{}
	tags := map[string]string{}

	for _, s := range b.Manifest.Services {
		tag := fmt.Sprintf("%s/%s", b.Options.Namespace, s.Name)

		if s.Image != "" {
			pulls[s.Image] = true
			tags[s.Image] = tag
		} else {
			builds[s.Build.Hash()] = s.Build
			tags[s.Build.Hash()] = tag
		}
	}

	// fmt.Printf("builds = %+v\n", builds)
	// fmt.Printf("pulls = %+v\n", pulls)
	// fmt.Printf("tags = %+v\n", tags)

	for _, build := range builds {
		if err := b.build(build); err != nil {
			return err
		}
	}

	for image := range pulls {
		if err := b.pull(image); err != nil {
			return err
		}
	}

	for from, to := range tags {
		if err := b.tag(from, to); err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) build(build manifest.ServiceBuild) error {
	tag := build.Hash()

	args := []string{"build"}

	// for _, arg := range build.Args {
	//   fmt.Printf("arg = %+v\n", arg)
	// }

	args = append(args, "-t", tag)

	path, err := b.Manifest.Path(build.Path)
	if err != nil {
		return err
	}

	args = append(args, path)

	b.message("building: %s", build.Path)

	if err := b.docker(args...); err != nil {
		return err
	}

	return nil
}

func (b *Builder) docker(args ...string) error {
	cmd := exec.Command("docker", args...)

	cmd.Stdout = b.stdout()
	cmd.Stderr = b.stderr()

	if _, err := cmd.Stdout.Write([]byte(fmt.Sprintf("running: docker %s\n", strings.Join(args, " ")))); err != nil {
		return err
	}

	return cmd.Run()
}

func (b *Builder) message(format string, args ...interface{}) {
	b.stdout().Write([]byte(fmt.Sprintf(format, args...) + "\n"))
}

func (b *Builder) pull(image string) error {
	b.message("pulling: %s", image)

	return b.docker("pull", image)
}

func (b *Builder) tag(from, to string) error {
	return b.docker("tag", from, to)
}

func (b *Builder) stderr() io.Writer {
	if b.Options.Stderr != nil {
		return b.Options.Stderr
	}

	return os.Stderr
}

func (b *Builder) stdout() io.Writer {
	if b.Options.Stdout != nil {
		return b.Options.Stdout
	}

	return os.Stdout
}
