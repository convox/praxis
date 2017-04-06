// +build darwin

package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/convox/praxis/types"
)

type plist struct {
	Action  string
	Exec    string
	Label   string
	LogPath string
}

const frontendPlist = "/Library/LaunchDaemons/com.convox.frontend.plist"

func (p *Provider) SystemInstall(name string, opts types.SystemInstallOptions) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	rackPlist := filepath.Join(u.HomeDir, "Library/LaunchAgents/com.convox.localrack.plist")

	t, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return "", err
	}

	// Create frontend daemone plist

	fep, err := ioutil.TempFile("", "frontend")
	if err != nil {
		return "", err
	}
	defer fep.Close()

	pl := plist{
		Action:  "frontend",
		Exec:    ex,
		Label:   "com.convox.frontend",
		LogPath: "/var/log",
	}

	if err := t.Execute(fep, pl); err != nil {
		return "", err
	}

	fmt.Println("You may be asked for your password")

	if err := sudoCmd("mv", fep.Name(), frontendPlist); err != nil {
		return "", err
	}

	if err := sudoCmd("chown", "root:wheel", frontendPlist); err != nil {
		return "", err
	}

	if err := sudoCmd("launchctl", "load", frontendPlist); err != nil {
		return "", err
	}

	// Create rack agent plist

	rp, err := os.Create(rackPlist)
	if err != nil {
		return "", err
	}
	defer rp.Close()

	pl = plist{
		Action:  "start",
		Exec:    ex,
		Label:   "com.convox.localrack",
		LogPath: p.Root,
	}

	if err := t.Execute(rp, pl); err != nil {
		return "", err
	}

	if out, err := exec.Command("launchctl", "load", rackPlist).CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to start local rack: %s - %s", err, string(out))
	}

	return "https://localhost:5443", nil
}

func (p *Provider) SystemUninstall(name string, opts types.SystemInstallOptions) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	rackPlist := filepath.Join(u.HomeDir, "Library/LaunchAgents/com.convox.localrack.plist")

	fmt.Println("You may be asked for your password")

	if err := sudoCmd("launchctl", "unload", frontendPlist); err != nil {
		return err
	}

	if err := sudoCmd("rm", frontendPlist); err != nil {
		return err
	}

	if out, err := exec.Command("launchctl", "unload", rackPlist).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop local rack: %s - %s", err, string(out))
	}

	if err := os.Remove(rackPlist); err != nil {
		return err
	}

	if err := os.RemoveAll(p.Root); err != nil {
		return err
	}

	return nil
}

const (
	plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Label</key>
		<string>{{ .Label }}</string>
		<key>ProgramArguments</key>
		<array>
			<string>{{ .Exec }}</string>
			<string>rack</string>
			<string>{{ .Action }}</string>
		</array>
                <key>KeepAlive</key>
		<true/>
                <key>StandardOutPath</key>
                <string>{{ .LogPath }}/{{ .Label }}.log</string>
                <key>StandardErrorPath</key>
                <string>{{ .LogPath }}/{{ .Label }}.log</string>
                <key>EnvironmentVariables</key>
                <dict>
	                <key>PATH</key>
	                <string>/sbin:/usr/sbin:/bin:/usr/bin:/usr/local/bin</string>
                </dict>
	</dict>
</plist>`
)
