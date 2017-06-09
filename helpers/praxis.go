package helpers

import (
	"fmt"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

func AppEnvironment(p types.Provider, app string) (types.Environment, error) {
	rs, err := p.ReleaseList(app, types.ReleaseListOptions{Count: 1})
	if err != nil {
		return nil, err
	}

	if len(rs) < 1 {
		return types.Environment{}, nil
	}

	return rs[0].Env, nil
}

func AppManifest(p types.Provider, app string) (*manifest.Manifest, *types.Release, error) {
	a, err := p.AppGet(app)
	if err != nil {
		return nil, nil, err
	}

	if a.Release == "" {
		return nil, nil, errors.WithStack(fmt.Errorf("no release for app: %s", app))
	}

	return ReleaseManifest(p, app, a.Release)
}

func ReleaseManifest(p types.Provider, app, release string) (*manifest.Manifest, *types.Release, error) {
	r, err := p.ReleaseGet(app, release)
	if err != nil {
		return nil, nil, err
	}

	if r.Build == "" {
		return nil, nil, errors.WithStack(fmt.Errorf("no builds for app: %s", app))
	}

	b, err := p.BuildGet(app, r.Build)
	if err != nil {
		return nil, nil, err
	}

	m, err := manifest.Load([]byte(b.Manifest), manifest.Environment(r.Env))
	if err != nil {
		return nil, nil, err
	}

	return m, r, nil
}
