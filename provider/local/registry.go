package local

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) RegistryAdd(server, username, password string) (*types.Registry, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) RegistryDelete(server string) error {
	return fmt.Errorf("unimplemented")
}

func (p *Provider) RegistryList() (types.Registries, error) {
	return nil, fmt.Errorf("unimplemented")
}
