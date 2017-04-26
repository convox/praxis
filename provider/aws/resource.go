package aws

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) ResourceCreate(name, kind string, params map[string]string) (*types.Resource, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) ResourceGet(name string) (*types.Resource, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) ResourceList() (types.Resources, error) {
	return nil, fmt.Errorf("unimplemented")
}
