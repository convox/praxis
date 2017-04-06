// +build linux

package local

import "github.com/convox/praxis/types"

type plist struct {
	Action  string
	Exec    string
	Label   string
	LogPath string
}

const frontendPlist = "/Library/LaunchDaemons/com.convox.frontend.plist"

func (p *Provider) SystemInstall(name string, opts types.SystemInstallOptions) (string, error) {
	return "https://localhost:5443", nil
}

func (p *Provider) SystemUninstall(name string, opts types.SystemInstallOptions) error {
	return nil
}
