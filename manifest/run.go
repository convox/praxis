package manifest

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/convox/praxis/fsync"
	"github.com/google/shlex"
)

type RunOptions struct {
	Sync bool
}

func (m *Manifest) Run(opts RunOptions) error {
	ch := make(chan error)

	for _, s := range m.Services {
		go m.runService(s, opts, ch)
	}

	for i := 0; i < len(m.Services); i++ {
		if err := <-ch; err != nil {
			return err
		}
	}

	return nil
}

func (m *Manifest) RunService(s Service, opts RunOptions) error {
	args := []string{"run"}

	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		args = append(args, "--env-file", ".env")
	}

	args = append(args, "-i")
	args = append(args, "--name", s.Name)
	args = append(args, "--rm")

	args = append(args, s.Name)

	if s.Command != "" {
		parts, err := shlex.Split(s.Command)
		if err != nil {
			return err
		}

		args = append(args, parts...)
	}

	if opts.Sync {
		go m.syncServicePaths(s)
	}

	m.system("starting: %s", s.Name)

	cmd := exec.Command("docker", args...)

	pw := m.prefix(s.Name)

	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (m *Manifest) runService(s Service, opts RunOptions, ch chan error) {
	ch <- m.RunService(s, opts)
}

func (m *Manifest) Stop() {
	var wg sync.WaitGroup

	wg.Add(len(m.Services))

	for _, s := range m.Services {
		m.system("stopping: %s", s.Name)

		go func(s Service, wg *sync.WaitGroup) {
			defer wg.Done()
			exec.Command("docker", "stop", "-t", "3", s.Name).Run()
		}(s, &wg)
	}

	wg.Wait()

	m.system("exit")
}

func (m *Manifest) syncServicePaths(s Service) {
	sp, err := s.SyncPaths()
	if err != nil {
		m.system("sync error: %s", err)
		return
	}

	for local, remote := range sp {
		s, err := fsync.NewSync(s.Name, local, remote)
		if err != nil {
			m.system("sync error: %s", err)
			return
		}

		ch := make(chan string)

		go s.Start(ch)

		for s := range ch {
			m.prefix("sync").Write([]byte(fmt.Sprintf("%s\n", s)))
		}
	}
}
