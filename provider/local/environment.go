package local

import (
	"fmt"
	"strings"

	"github.com/convox/praxis/types"
)

func (p *Provider) EnvironmentGet(app string) (types.Environment, error) {
	var env types.Environment

	if err := p.storageLoad(fmt.Sprintf("apps/%s/environment", app), &env); err != nil {
		if strings.HasPrefix(err.Error(), "no such key:") {
			return types.Environment{}, nil
		}
		return nil, err
	}

	return env, nil
}

func (p *Provider) EnvironmentSet(app string, env types.Environment) error {
	cenv, err := p.EnvironmentGet(app)
	if err != nil {
		return err
	}

	for k, v := range env {
		cenv[k] = v
	}

	return p.storageStore(fmt.Sprintf("apps/%s/environment", app), cenv)
}

func (p *Provider) EnvironmentUnset(app string, key string) error {
	cenv, err := p.EnvironmentGet(app)
	if err != nil {
		return err
	}

	delete(cenv, key)

	return p.storageStore(fmt.Sprintf("apps/%s/environment", app), cenv)
}
