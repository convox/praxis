package aws

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/convox/praxis/types"
)

func (p *Provider) EnvironmentGet(app string) (types.Environment, error) {
	exists, err := p.ObjectExists(app, "convox/environment")
	if err != nil {
		return nil, err
	}

	if !exists {
		return types.Environment{}, nil
	}

	r, err := p.ObjectFetch(app, "convox/environment")
	if err != nil {
		return nil, err
	}

	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var env types.Environment

	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}

	return env, nil
}

func (p *Provider) EnvironmentSet(app string, env types.Environment) error {
	eenv, err := p.EnvironmentGet(app)
	if err != nil {
		return err
	}

	for k, v := range env {
		eenv[k] = v
	}

	data, err := json.Marshal(eenv)
	if err != nil {
		return err
	}

	if _, err := p.ObjectStore(app, "convox/environment", bytes.NewReader(data), types.ObjectStoreOptions{}); err != nil {
		return err
	}

	return nil
}

func (p *Provider) EnvironmentUnset(app string, key string) error {
	env, err := p.EnvironmentGet(app)
	if err != nil {
		return err
	}

	delete(env, key)

	return p.EnvironmentSet(app, env)
}
