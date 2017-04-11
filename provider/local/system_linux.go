package local

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func launcherPath(name string, opts launchOptions) string {
	return filepath.Join("/lib/systemd/system", fmt.Sprintf("%s.service", name))
}

func launcherStart(name string, opts launchOptions) error {
	return exec.Command("systemctl", "start", name).Run()
}

func launcherStop(name string, opts launchOptions) error {
	return exec.Command("systemctl", "stop", name).Run()
}

func launcherTemplate() string {
	return `
[Unit]
After=network.target

[Service]
Type=simple
ExecStart={{ .Executable }} {{ range .Args }}{{ . }} {{ end }}
KillMode=control-group
Restart=always
RestartSec=10s

[Install]
WantedBy=default.target`
}
