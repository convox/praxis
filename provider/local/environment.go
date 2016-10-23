package local

import (
	"encoding/json"
	"io/ioutil"

	"github.com/convox/praxis/provider"
)

func (p *Provider) EnvironmentLoad(app string) (provider.Environment, error) {
	exists, err := p.BlobExists(app, "env")
	if err != nil {
		return nil, err
	}

	var env provider.Environment

	if exists {
		r, err := p.BlobFetch(app, "env")
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &env); err != nil {
			return nil, err
		}
	}

	return env, nil
}

func (p *Provider) EnvironmentSave(app string, env provider.Environment) error {
	return nil
}
