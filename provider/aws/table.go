package aws

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) TableCreate(app, name string, opts types.TableCreateOptions) error {
	return fmt.Errorf("unimplemented")
}

func (p *Provider) TableGet(app, table string) (*types.Table, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) TableList(app string) (types.Tables, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) TableQuery(app, table, query string) (types.TableRows, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) TableTruncate(app, table string) error {
	return fmt.Errorf("unimplemented")
}
