package local

import (
	"fmt"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func (p *Provider) TableFetch(app, table, id string) (attrs map[string]string, err error) {
	err = p.Load(fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s/%s.json", app, table, "id", id, id), &attrs)
	return
}

func (p *Provider) TableFetchIndex(app, table, index, key string) ([]map[string]string, error) {
	files, err := p.List(fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s", app, table, index, key))
	if err != nil {
		return nil, err
	}

	var items []map[string]string

	for _, f := range files {
		attrs := map[string]string{}
		if err := p.Load(fmt.Sprintf("apps/%s/tables/%s/indexes/%s/%s/%s", app, table, index, key, f), &attrs); err != nil {
			return nil, err
		}

		items = append(items, attrs)
	}

	return items, nil
}

func (p *Provider) TableFetchIndexBatch(app, table, index string, keys []string) ([]map[string]string, error) {
	var batch []map[string]string

	for _, k := range keys {
		attrs, err := p.TableFetchIndex(app, table, index, k)
		if err != nil {
			return nil, err
		}

		batch = append(batch, attrs...)
	}

	return batch, nil
}

func (p *Provider) TableGet(app, table string) (*manifest.Table, error) {
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
			return &t, nil
		}
	}

	return nil, fmt.Errorf("table %s not defined", table)
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

	key := "apps/%s/tables/%s/indexes/%s/%s/%s.json"
	var idFound bool
	for _, i := range t.Indexes {
		if attrs[i] != "" {
			if err := p.Store(fmt.Sprintf(key, app, table, i, attrs[i], attrs["id"]), attrs); err != nil {
				return "", err
			}
		}

		if i == "id" {
			idFound = true
		}
	}

	if !idFound {
		if err := p.Store(fmt.Sprintf(key, app, table, "id", attrs["id"], attrs["id"]), attrs); err != nil {
			return "", err
		}
	}

	return attrs["id"], nil
}
