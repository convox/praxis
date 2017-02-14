package local

import (
	"io"

	"github.com/convox/praxis/provider/types"
)

func (p *Provider) ObjectStore(app, key string, r io.Reader, opts types.ObjectStoreOptions) (*types.Object, error) {
	return &types.Object{Key: key}, nil
}
