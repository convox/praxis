package helpers

import (
	"fmt"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func AppManifest(p types.Provider, app string) (*manifest.Manifest, *types.Release, error) {
	a, err := p.AppGet(app)
	if err != nil {
		return nil, nil, err
	}

	if a.Release == "" {
		return nil, nil, fmt.Errorf("no releases for app: %s", app)
	}

	r, err := p.ReleaseGet(app, a.Release)
	if err != nil {
		return nil, nil, err
	}

	if r.Build == "" {
		return nil, nil, fmt.Errorf("no builds for app: %s", app)
	}

	b, err := p.BuildGet(app, r.Build)
	if err != nil {
		return nil, nil, err
	}

	m, err := manifest.Load([]byte(b.Manifest), r.Env)
	if err != nil {
		return nil, nil, err
	}

	return m, r, nil
}
