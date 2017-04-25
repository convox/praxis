package local

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/convox/praxis/types"
)

func (p *Provider) ObjectExists(app, key string) (bool, error) {
	if _, err := p.AppGet(app); err != nil {
		return false, err
	}

	token := fmt.Sprintf("apps/%s/objects/%s", app, key)

	return p.storageExists(token), nil
}

func (p *Provider) ObjectFetch(app, key string) (io.ReadCloser, error) {
	if _, err := p.AppGet(app); err != nil {
		return nil, err
	}

	token := fmt.Sprintf("apps/%s/objects/%s", app, key)

	if !p.storageExists(token) {
		return nil, fmt.Errorf("no such key: %s", key)
	}

	data, err := p.storageRead(token)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(bytes.NewReader(data)), nil
}

func (p *Provider) ObjectStore(app, key string, r io.Reader, opts types.ObjectStoreOptions) (*types.Object, error) {
	if _, err := p.AppGet(app); err != nil {
		return nil, err
	}

	if key == "" {
		return nil, fmt.Errorf("key must not be blank")
	}

	if err := p.storageStore(fmt.Sprintf("apps/%s/objects/%s", app, key), r); err != nil {
		return nil, err
	}

	return &types.Object{Key: key}, nil
}
