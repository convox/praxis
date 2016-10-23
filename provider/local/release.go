package local

import (
	"encoding/json"

	"github.com/convox/praxis/provider"
)

func (p *Provider) ReleaseCreate(app string, build *provider.Build, env provider.Environment) (*provider.Release, error) {
	release := &provider.Release{
		Id:          id("R"),
		Build:       build.Id,
		Environment: env,
	}

	if err := p.ReleaseSave(release); err != nil {
		return nil, err
	}

	return release, nil
}

func (p *Provider) ReleaseSave(release *provider.Release) error {
	env, err := json.Marshal(release.Environment)
	if err != nil {
		return err
	}

	attrs := map[string]string{
		"build":       release.Build,
		"environment": string(env),
	}

	return p.TableSave("system", "releases", release.Id, attrs)
}
