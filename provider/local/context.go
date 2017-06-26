package local

import (
	"context"

	"github.com/convox/praxis/types"
)

func (p *Provider) WithContext(ctx context.Context) types.Provider {
	var q Provider
	q = *p
	q.ctx = ctx
	return &q
}
