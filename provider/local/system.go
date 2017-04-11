package local

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/convox/praxis/types"
	homedir "github.com/mitchellh/go-homedir"
)

var (
	launcher = template.Must(template.New("launcher").Parse(launcherTemplate()))
)

type launchOptions struct {
	Args []string
	Sudo bool
}

func (p *Provider) SystemGet() (*types.System, error) {
	system := &types.System{
		Name:    "convox",
		Image:   fmt.Sprintf("convox/praxis:%s", os.Getenv("VERSION")),
		Version: os.Getenv("VERSION"),
	}

	return system, nil
}

func (p *Provider) SystemInstall(name string, opts types.SystemInstallOptions) (string, error) {
	err := launcherInstall("convox.frontend", launchOptions{
		Args: []string{"rack", "frontend"},
		Sudo: true,
	})
	if err != nil {
		return "", err
	}

	err = launcherInstall("convox.rack", launchOptions{
		Args: []string{"rack", "start"},
	})
	if err != nil {
		return "", err
	}

	return "https://localhost:5443", nil
}

func (p *Provider) SystemUninstall(name string, opts types.SystemInstallOptions) error {
	launcherRemove("convox.frontend", launchOptions{Sudo: true})
	launcherRemove("convox.rack", launchOptions{})

	return nil
}

func launcherInstall(name string, opts launchOptions) error {
	var buf bytes.Buffer

	ex, err := os.Executable()
	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"Executable": ex,
		"Name":       name,
		"Args":       opts.Args,
		"Logs":       fmt.Sprintf("/var/log/%s.log", name),
	}

	path := launcherPath(name, opts)

	if !opts.Sudo {
		u, err := user.Current()
		if err != nil {
			return err
		}

		h, err := homedir.Dir()
		if err != nil {
			return err
		}

		params["Logs"] = filepath.Join(h, ".convox", "local", "rack.log")
		params["User"] = u.Username
	}

	if err := launcher.Execute(&buf, params); err != nil {
		return err
	}

	fmt.Printf("installing: %s\n", path)

	if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return err
	}

	if err := launcherStart(name, opts); err != nil {
		return err
	}

	return nil
}

func launcherRemove(name string, opts launchOptions) error {
	path := launcherPath(name, opts)

	fmt.Printf("removing: %s\n", path)

	launcherStop(name, opts)
	os.Remove(path)

	return nil
}
