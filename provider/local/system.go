package local

import (
	"fmt"
	"os"

	"github.com/convox/praxis/types"
)

func (p *Provider) SystemGet() (*types.System, error) {
	system := &types.System{
		Name:    "convox",
		Image:   fmt.Sprintf("convox/praxis:%s", os.Getenv("VERSION")),
		Version: os.Getenv("VERSION"),
	}

	return system, nil
}

func (p *Provider) SystemInstall(name string, opts types.SystemInstallOptions) (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (p *Provider) SystemUninstall(name string, opts types.SystemInstallOptions) error {
	return fmt.Errorf("unimplemented")
}
