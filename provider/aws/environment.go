package aws

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) EnvironmentDelete(app string, key string) error {
	return fmt.Errorf("unimplemented")
}

func (p *Provider) EnvironmentGet(app string) (types.Environment, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) EnvironmentSet(app string, env types.Environment) error {
	return fmt.Errorf("unimplemented")
}
