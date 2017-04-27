package local

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/convox/praxis/types"
)

var (
	launcher = template.Must(template.New("launcher").Parse(launcherTemplate()))
)

func (p *Provider) SystemGet() (*types.System, error) {
	system := &types.System{
		Domain:  p.Name,
		Name:    p.Name,
		Image:   fmt.Sprintf("convox/praxis:%s", p.Version),
		Version: p.Version,
	}

	return system, nil
}

func (p *Provider) SystemInstall(name string, opts types.SystemInstallOptions) (string, error) {
	cx, err := os.Executable()
	if err != nil {
		return "", err
	}

	if err := launcherInstall("convox.frontend", cx, "rack", "frontend"); err != nil {
		return "", err
	}

	if err := launcherInstall("convox.rack", cx, "rack", "start"); err != nil {
		return "", err
	}

	return "https://localhost:5443", nil
}

func (p *Provider) SystemUninstall(name string, opts types.SystemInstallOptions) error {
	launcherRemove("convox.frontend")
	launcherRemove("convox.rack")

	return nil
}

func (p *Provider) SystemUpdate(opts types.SystemUpdateOptions) error {
	w := opts.Output
	if w == nil {
		w = ioutil.Discard
	}

	if v := opts.Version; v != "" {
		w.Write([]byte("Restarting... OK\n"))

		if err := ioutil.WriteFile("/var/convox/version", []byte(v), 0644); err != nil {
			return err
		}

		defer os.Exit(0)
	}

	return nil
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
