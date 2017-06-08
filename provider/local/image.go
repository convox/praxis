package local

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) ImageCreate(name, url string, opts types.ImageCreateOptions) (*types.Image, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) ImageList() (types.Images, error) {
	return nil, fmt.Errorf("unimplemented")
}
