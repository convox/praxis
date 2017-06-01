package local

import (
	"fmt"

	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

func (p *Provider) RegistryAdd(hostname, username, password string) (*types.Registry, error) {
	log := p.logger("RegistryAdd").Append("hostname=%q username=%q", hostname, username)

	r := &types.Registry{
		Hostname: hostname,
		Username: username,
		Password: password,
	}

	key := fmt.Sprintf("registries/%s", hostname)

	if p.storageExists(key) {
		return nil, log.Error(fmt.Errorf("registry already exists: %s", hostname))
	}

	if err := p.storageStore(fmt.Sprintf("registries/%s", hostname), r); err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	return r, log.Success()
}

func (p *Provider) RegistryList() (types.Registries, error) {
	log := p.logger("RegistryList")

	names, err := p.storageList("registries")
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	registries := make(types.Registries, len(names))

	var r types.Registry

	for i, name := range names {
		if err := p.storageLoad(fmt.Sprintf("registries/%s", name), &r); err != nil {
			return nil, errors.WithStack(log.Error(err))
		}

		registries[i] = r
	}

	return registries, log.Success()
}

func (p *Provider) RegistryRemove(hostname string) error {
	log := p.logger("RegistryAdd").Append("hostname=%q", hostname)

	key := fmt.Sprintf("registries/%s", hostname)

	if !p.storageExists(key) {
		return log.Error(fmt.Errorf("no such registry: %s", hostname))
	}

	if err := p.storageDelete(key); err != nil {
		errors.WithStack(log.Error(err))
	}

	return log.Success()
}
