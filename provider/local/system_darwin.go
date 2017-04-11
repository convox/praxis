package local

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func launcherPath(name string, opts launchOptions) string {
	if opts.Sudo {
		return filepath.Join("/Library/LaunchDaemons", fmt.Sprintf("%s.plist", name))
	} else {
		return filepath.Join("/Library/LaunchAgents", fmt.Sprintf("%s.plist", name))
	}
}

func launcherStart(name string, opts launchOptions) error {
	return exec.Command("launchctl", "load", launcherPath(name, opts)).Run()
}

func launcherStop(name string, opts launchOptions) error {
	return exec.Command("launchctl", "remove", name).Run()
}

func launcherTemplate() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
		<dict>
			<key>Label</key>
			<string>{{ .Name }}</string>
			<key>ProgramArguments</key>
			<array>
				<string>{{ .Executable }}</string>
				{{ range .Args }}
					<string>{{ . }}</string>
				{{ end }}
			</array>
			<key>EnvironmentVariables</key>
			<dict>
				<key>PATH</key>
				<string>/sbin:/usr/sbin:/bin:/usr/bin:/usr/local/bin</string>
			</dict>
			<key>RunAtLoad</key>
			<true/>
			<key>KeepAlive</key>
			<true/>
			{{ if .User }}
				<key>UserName</key>
				<string>{{ .User }}</string>
			{{ end }}
			{{ if .Logs }}
				<key>StandardOutPath</key>
				<string>{{ .Logs }}</string>
				<key>StandardErrorPath</key>
				<string>{{ .Logs }}</string>
			{{ end }}
		</dict>
	</plist>`
}
