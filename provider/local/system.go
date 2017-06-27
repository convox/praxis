package local

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/user"
	"text/template"
	"time"

	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

var (
	launcher = template.Must(template.New("launcher").Parse(launcherTemplate()))
)

func (p *Provider) SystemGet() (*types.System, error) {
	log := p.logger("SystemGet")

	system := &types.System{
		Image:   fmt.Sprintf("convox/praxis:%s", p.Version),
		Name:    p.Name,
		Status:  "running",
		Version: p.Version,
	}

	return system, log.Success()
}

func (p *Provider) SystemInstall(name string, opts types.SystemInstallOptions) (string, error) {
	cx, err := os.Executable()
	if err != nil {
		return "", err
	}

	u, err := user.Current()
	if err != nil {
		return "", err
	}

	if u.Uid != "0" {
		return "", fmt.Errorf("must be root to install a local rack")
	}

	if err := launcherInstall("convox.router", cx, "router"); err != nil {
		return "", err
	}

	if err := launcherInstall("convox.rack", cx, "rack", "start"); err != nil {
		return "", err
	}

	return "https://localhost:5443", nil
}

func (p *Provider) SystemLogs(opts types.LogsOptions) (io.ReadCloser, error) {
	log := p.logger("SystemLogs")

	r, w := io.Pipe()

	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	args := []string{"logs"}

	if opts.Follow {
		args = append(args, "-f")
	}

	if !opts.Since.IsZero() {
		args = append(args, "--since", opts.Since.Format(time.RFC3339))
	}

	args = append(args, hostname)

	cmd := exec.Command("docker", args...)

	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Start(); err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	go func() {
		defer w.Close()
		cmd.Wait()
	}()

	return r, log.Success()
}

func (p *Provider) SystemOptions() (map[string]string, error) {
	log := p.logger("SystemOptions")

	options := map[string]string{
		"streaming": "http2",
	}

	return options, log.Success()
}

func (p *Provider) SystemProxy(host string, port int, in io.Reader) (io.ReadCloser, error) {
	log := p.logger("SystemProxy").Append("host=%s port=%d", host, port)

	cn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	go io.Copy(cn, in)

	return cn, log.Success()
}

func (p *Provider) SystemUninstall(name string, opts types.SystemInstallOptions) error {
	launcherRemove("convox.frontend")
	launcherRemove("convox.rack")
	launcherRemove("convox.router")

	exec.Command("launchctl", "remove", "convox.frontend").Run()
	exec.Command("launchctl", "remove", "convox.rack").Run()
	exec.Command("launchctl", "remove", "convox.router").Run()

	return nil
}

func (p *Provider) SystemUpdate(opts types.SystemUpdateOptions) error {
	log := p.logger("SystemUpdate").Append("version=%q", opts.Version)

	w := opts.Output
	if w == nil {
		w = ioutil.Discard
	}

	if v := opts.Version; v != "" {
		w.Write([]byte("Restarting... OK\n"))

		if err := ioutil.WriteFile("/var/convox/version", []byte(v), 0644); err != nil {
			return errors.WithStack(log.Error(err))
		}

		if err := exec.Command("docker", "pull", fmt.Sprintf("convox/praxis:%s", v)).Run(); err != nil {
			return errors.WithStack(log.Error(err))
		}

		go func() {
			time.Sleep(1 * time.Second)
			os.Exit(0)
		}()
	}

	return log.Success()
}

func launcherInstall(name string, command string, args ...string) error {
	var buf bytes.Buffer

	params := map[string]interface{}{
		"Name":    name,
		"Command": command,
		"Args":    args,
		"Logs":    fmt.Sprintf("/var/log/%s.log", name),
	}

	if err := launcher.Execute(&buf, params); err != nil {
		return err
	}

	path := launcherPath(name)

	fmt.Printf("installing: %s\n", path)

	if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return err
	}

	if err := launcherStart(name); err != nil {
		return err
	}

	return nil
}

func launcherRemove(name string) error {
	path := launcherPath(name)

	fmt.Printf("removing: %s\n", path)

	launcherStop(name)

	os.Remove(path)

	return nil
}
