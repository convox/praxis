package aws

import (
	"github.com/convox/praxis/types"
)

func (p *Provider) TableCreate(app, name string, opts types.TableCreateOptions) error {
	return nil
}

func (p *Provider) TableGet(app, table string) (*types.Table, error) {
	return nil, nil
}

func (p *Provider) TableList(app string) (types.Tables, error) {
	return nil, nil
}

func (p *Provider) TableTruncate(app, table string) error {
	return nil
}

func (p *Provider) TableRowDelete(app, table, key string, opts types.TableRowDeleteOptions) error {
	return nil
}

func (p *Provider) TableRowGet(app, table, key string, opts types.TableRowGetOptions) (*types.TableRow, error) {
	return nil, nil
}

func (p *Provider) TableRowStore(app, table string, attrs types.TableRow) (string, error) {
	return "", nil
}

func (p *Provider) TableRowsDelete(app, table string, keys []string, opts types.TableRowDeleteOptions) error {
	return nil
}

func (p *Provider) TableRowsGet(app, table string, keys []string, opts types.TableRowGetOptions) (types.TableRows, error) {
	return nil, nil
}
