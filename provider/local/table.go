package local

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/convox/praxis/types"
)

const (
	TableCacheDuration = 5 * time.Minute
)

func (p *Provider) TableGet(app, table string) (*types.Table, error) {
	var t *types.Table

	if err := p.storageLoad(fmt.Sprintf("apps/%s/tables/%s/table.json", app, table), &t, TableCacheDuration); err != nil {
		if strings.HasPrefix(err.Error(), "no such key:") {
			return nil, fmt.Errorf("no such table: %s", table)
		}
		return nil, err
	}

	return t, nil
}

func (p *Provider) TableList(app string) (types.Tables, error) {
	tt, err := p.storageList(fmt.Sprintf("apps/%s/tables", app))
	if err != nil {
		return nil, err
	}

	tables := make(types.Tables, len(tt))

	for i, t := range tt {
		table, err := p.TableGet(app, t)
		if err != nil {
			return nil, err
		}

		tables[i] = *table
	}

	sort.Slice(tables, tables.Less)

	return tables, nil
}

func (p *Provider) TableQuery(app, table, query string) (types.TableRows, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) TableTruncate(app, table string) error {
	return p.storageDeleteAll(fmt.Sprintf("apps/%s/tables/%s/indexes", app, table))
}
