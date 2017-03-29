package local

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) RegistryAdd(hostname, username, password string) (*types.Registry, error) {
	r := &types.Registry{
		Hostname: hostname,
		Username: username,
		Password: password,
	}

	key := fmt.Sprintf("registries/%s", hostname)

	if p.storageExists(key) {
		return nil, fmt.Errorf("registry already exists: %s", hostname)
	}

	if err := p.storageStore(fmt.Sprintf("registries/%s", hostname), r); err != nil {
		return nil, err
	}

	return r, nil
}

func (p *Provider) RegistryList() (types.Registries, error) {
	names, err := p.storageList("registries")
	if err != nil {
		return nil, err
	}

	registries := make(types.Registries, len(names))

	var r types.Registry

	for i, name := range names {
		if err := p.storageLoad(fmt.Sprintf("registries/%s", name), &r); err != nil {
			return nil, err
		}

		registries[i] = r
	}

	return registries, nil
}

func (p *Provider) RegistryRemove(hostname string) error {
	key := fmt.Sprintf("registries/%s", hostname)

	if !p.storageExists(key) {
		return fmt.Errorf("no such registry: %s", hostname)
	}

	return p.storageDelete(key)
}
