package local

import (
	"fmt"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func (p *Provider) TableFetch(app, table, key string, opts types.TableFetchOptions) (map[string]string, error) {
	items, err := p.TableFetchBatch(app, table, []string{key}, opts)
	if err != nil {
		return nil, err
	}

	switch len(items) {
	case 0:
		return nil, fmt.Errorf("not found")
	case 1:
		return items[0], nil
	default:
		return nil, fmt.Errorf("multiple items found")
	}
}

func (p *Provider) TableFetchBatch(app, table string, keys []string, opts types.TableFetchOptions) ([]map[string]string, error) {
	items := []map[string]string{}

	for _, key := range keys {
		if key == "" {
			continue
		}

		entries, err := p.List(fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s/", app, table, coalesce(opts.Index, "id"), key))
		if err != nil {
			return nil, err
		}

		for _, e := range entries {
			var item map[string]string

			if err := p.Load(fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s/%s", app, table, coalesce(opts.Index, "id"), key, e), &item); err != nil {
				return nil, err
			}

			items = append(items, item)
		}
	}

	return items, nil
}

func (p *Provider) TableGet(app, table string) (*types.Table, error) {
	releases, err := p.ReleaseList(app)
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
	}

	build, err := p.BuildGet(app, releases[0].Build)
	if err != nil {
		return nil, err
	}

	m, err := manifest.Load([]byte(build.Manifest))
	if err != nil {
		return nil, err
	}

	for _, t := range m.Tables {
		if t.Name == table {
			return &types.Table{
				Name:    t.Name,
				Indexes: t.Indexes,
			}, nil
		}
	}

	return nil, fmt.Errorf("no such table: %s", table)
}

func (p *Provider) TableList(app string) (types.Tables, error) {
	tt, err := p.List(fmt.Sprintf("apps/%s/tables", app))
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

	return tables, nil
}

func (p *Provider) TableStore(app, table string, attrs map[string]string) (string, error) {
	if attrs["id"] == "" {
		id, err := types.Key(64)
		if err != nil {
			return "", err
		}
		attrs["id"] = id
	}

	t, err := p.TableGet(app, table)
	if err != nil {
		return "", err
	}

	indexes := append(t.Indexes, "id")

	for _, index := range indexes {
		if attrs[index] == "" {
			continue
		}

		if err := p.Store(fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s/%s.json", app, table, index, attrs[index], attrs["id"]), attrs); err != nil {
			return "", err
		}
	}

	return attrs["id"], nil
}

func (p *Provider) TableTruncate(app, table string) error {
	return p.DeleteAll(fmt.Sprintf("apps/%s/tables/%s", app, table))
}
