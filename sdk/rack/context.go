package rack

import (
	"context"

	"github.com/convox/praxis/types"
)

func (c *Client) WithContext(ctx context.Context) types.Provider {
	return c
}
