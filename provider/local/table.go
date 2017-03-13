package local

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/convox/praxis/types"
)

func (p *Provider) TableCreate(app, name string, opts types.TableCreateOptions) error {
	t := types.Table{
		Name:    name,
		Indexes: opts.Indexes,
	}

	return p.Store(fmt.Sprintf("apps/%s/tables/%s/table.json", app, name), t)
}

func (p *Provider) TableGet(app, table string) (*types.Table, error) {
	var t *types.Table

	if err := p.Load(fmt.Sprintf("apps/%s/tables/%s/table.json", app, table), &t); err != nil {
		if strings.HasPrefix(err.Error(), "no such key:") {
			return nil, fmt.Errorf("no such table: %s", table)
		}
		return nil, err
	}

	return t, nil
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

func (p *Provider) TableTruncate(app, table string) error {
	return p.DeleteAll(fmt.Sprintf("apps/%s/tables/%s/indexes", app, table))
}

func (p *Provider) TableRowDelete(app, table, key string, opts types.TableRowDeleteOptions) error {
	return p.TableRowsDelete(app, table, []string{key}, opts)
}

func (p *Provider) TableRowGet(app, table, key string, opts types.TableRowGetOptions) (*types.TableRow, error) {
	items, err := p.TableRowsGet(app, table, []string{key}, opts)
	if err != nil {
		return nil, err
	}

	switch len(items) {
	case 0:
		return nil, fmt.Errorf("not found")
	case 1:
		return &items[0], nil
	default:
		return nil, fmt.Errorf("multiple items found")
	}
}

func (p *Provider) TableRowStore(app, table string, attrs types.TableRow) (string, error) {
	if attrs["id"] == "" {
		id, err := types.Key(64)
		if err != nil {
			return "", err
		}
		attrs["id"] = id
	} else {
		row, err := p.TableRowGet(app, table, attrs["id"], types.TableRowGetOptions{})
		if err != nil {
			return "", err
		}

		for k := range attrs {
			(*row)[k] = attrs[k]
		}

		attrs = *row

		if err := p.TableRowDelete(app, table, attrs["id"], types.TableRowDeleteOptions{}); err != nil {
			return "", err
		}
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

		ea := encodeAttrs(attrs)

		if err := p.Store(fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s/%s.json", app, table, index, ea[index], ea["id"]), ea); err != nil {
			return "", err
		}
	}

	return attrs["id"], nil
}

func (p *Provider) TableRowsDelete(app, table string, keys []string, opts types.TableRowDeleteOptions) error {
	t, err := p.TableGet(app, table)
	if err != nil {
		return err
	}
	indexes := append(t.Indexes, "id")

	items, err := p.TableRowsGet(app, table, keys, types.TableRowGetOptions{Index: opts.Index})
	if err != nil {
		return err
	}

	for _, item := range items {
		for _, in := range indexes {
			if item[in] == "" {
				continue
			}

			ei := encodeAttrs(item)

			if err := p.Delete(fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s/%s.json", app, table, in, ei[in], ei["id"])); err != nil {
				return err
			}

			dir := fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s/", app, table, in, ei[in])
			entries, err := p.List(dir)
			if err != nil {
				return err
			}

			if len(entries) == 0 {
				if err := p.Delete(dir); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (p *Provider) TableRowsGet(app, table string, keys []string, opts types.TableRowGetOptions) (types.TableRows, error) {
	items := types.TableRows{}

	for _, key := range keys {
		if key == "" {
			continue
		}

		ek := encodeValue(key)

		entries, err := p.List(fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s/", app, table, coalesce(opts.Index, "id"), ek))
		if err != nil {
			return nil, err
		}

		for _, e := range entries {
			var attrs map[string]string

			if err := p.Load(fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s/%s", app, table, coalesce(opts.Index, "id"), ek, e), &attrs); err != nil {
				return nil, err
			}

			da, err := decodeAttrs(attrs)
			if err != nil {
				return nil, err
			}

			items = append(items, da)
		}
	}

	return items, nil
}

func decodeAttrs(attrs map[string]string) (map[string]string, error) {
	dec := map[string]string{}

	for k, v := range attrs {
		d, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}

		dec[k] = string(d)
	}

	return dec, nil
}

func encodeAttrs(attrs map[string]string) map[string]string {
	enc := map[string]string{}

	for k, v := range attrs {
		enc[k] = encodeValue(v)
	}

	return enc
}

func encodeValue(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
