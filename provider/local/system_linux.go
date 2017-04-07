// +build linux

package local

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/convox/praxis/types"
)

type unit struct {
	Description string
	Exec        string
	Target      string
}

var (
	frontendService = "convox-frontend.service"
	rackService     = "convox-local-rack.service"

	// path works for fedora and debian
	frontendUnit = filepath.Join("/lib/systemd/system", frontendService)
)

func (p *Provider) SystemInstall(name string, opts types.SystemInstallOptions) (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	t, err := template.New("unitTemplate").Parse(unitTemplate)
	if err != nil {
		return "", err
	}

	// Create frontend service

	feu, err := ioutil.TempFile("", "frontend")
	if err != nil {
		return "", err
	}
	defer feu.Close()

	pl := unit{
		Description: "Run a the local frontend",
		Exec:        fmt.Sprintf("%s rack frontend", ex),
		Target:      "multi-user.target",
	}

	if err := t.Execute(feu, pl); err != nil {
		return "", err
	}

	if err := sudoCmd("mv", feu.Name(), frontendUnit); err != nil {
		return "", err
	}

	if err := sudoCmd("chown", "root:root", frontendUnit); err != nil {
		return "", err
	}

	if err := sudoCmd("chmod", "644", frontendUnit); err != nil {
		return "", err
	}

	if err := sudoCmd("systemctl", "enable", frontendService); err != nil {
		return "", err
	}

	if err := sudoCmd("systemctl", "start", frontendService); err != nil {
		return "", err
	}

	// Create local rack service

	u, err := user.Current()
	if err != nil {
		return "", err
	}

	rackUnit := filepath.Join(systemdUserDir(u), rackService)

	lru, err := os.Create(rackUnit)
	if err != nil {
		return "", err
	}
	defer lru.Close()

	pl = unit{
		Description: "Run a local rack",
		Exec:        fmt.Sprintf("%s rack start", ex),
		Target:      "default.target",
	}

	if err := t.Execute(lru, pl); err != nil {
		return "", err
	}

	if err := sudoCmd("loginctl", "enable-linger", u.Name); err != nil {
		return "", err
	}

	if out, err := exec.Command("systemctl", "--user", "enable", rackService).CombinedOutput(); err != nil {
		return "", fmt.Errorf("service enable failed: %s - %s", strings.TrimSpace(string(out)), err)
	}

	if out, err := exec.Command("systemctl", "--user", "start", rackService).CombinedOutput(); err != nil {
		return "", fmt.Errorf("service start failed: %s - %s", strings.TrimSpace(string(out)), err)
	}

	return "https://localhost:5443", nil
}

func (p *Provider) SystemUninstall(name string, opts types.SystemInstallOptions) error {
	return nil
}

func systemdUserDir(u *user.User) string {
	homeCfg := os.Getenv("$XDG_CONFIG_HOME")
	if homeCfg != "" {
		return filepath.Join(homeCfg, "systemd/user")
	}

	return filepath.Join(u.HomeDir, "/.config/systemd/user")
}

const unitTemplate = `[Unit]
Description={{ .Description }}
After=network.target

[Service]
Type=simple
ExecStart={{ .Exec }}
KillMode=control-group
Restart=always
RestartSec=10s

[Install]
WantedBy={{ .Target }}`
