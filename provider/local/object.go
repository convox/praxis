package local

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/convox/praxis/types"
	"github.com/pkg/errors"
)

func (p *Provider) ObjectExists(app, key string) (bool, error) {
	log := p.logger("ObjectExists").Append("app=%s key=%q", app, key)

	if _, err := p.AppGet(app); err != nil {
		return false, err
	}

	fn := filepath.Join(p.Root, "apps", app, "objects", key)

	_, err := os.Stat(fn)

	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return false, errors.WithStack(log.Error(err))
	}

	return true, log.Success()
}

func (p *Provider) ObjectFetch(app, key string) (io.ReadCloser, error) {
	if _, err := p.AppGet(app); err != nil {
		return nil, err
	}

	ex, err := p.ObjectExists(app, key)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("no such key: %s", key)
	}

	fn := filepath.Join(p.Root, "apps", app, "objects", key)

	fd, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	return fd, nil
}

func (p *Provider) ObjectStore(app, key string, r io.Reader, opts types.ObjectStoreOptions) (*types.Object, error) {
	log := p.logger("ObjectStore").Append("app=%s key=%q", app, key)

	if _, err := p.AppGet(app); err != nil {
		return nil, err
	}

	if key == "" {
		return nil, fmt.Errorf("key must not be blank")
	}

	fn := filepath.Join(p.Root, "apps", app, "objects", key)

	dir := filepath.Dir(fn)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	fd, err := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	defer fd.Close()

	if _, err := io.Copy(fd, r); err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	return &types.Object{Key: key}, log.Success()
}
