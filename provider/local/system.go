package local

import (
	"fmt"
	"os"

	"github.com/convox/praxis/types"
)

func (p *Provider) SystemGet() (*types.System, error) {
	system := &types.System{
		Name:    "convox",
		Image:   fmt.Sprintf("convox/praxis:%s", os.Getenv("VERSION")),
		Version: os.Getenv("VERSION"),
	}

	return system, nil
}
