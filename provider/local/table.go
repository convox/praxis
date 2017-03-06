package local

import (
	"fmt"

	"github.com/convox/praxis/types"
)

func (p *Provider) TableFetch(table string, id string) (attrs map[string]string, err error) {
	err = p.Load(fmt.Sprintf("tables/%s/indexes/id/%s/%s.json", table, id, id), &attrs)
	return
}

func (p *Provider) TableStore(table string, attrs map[string]string) (string, error) {
	id, err := types.Key(64)
	if err != nil {
		return "", err
	}

	if attrs["id"] == "" {
		attrs["id"] = id
	}

	indexes := []string{"id", "name"}
	for _, i := range indexes {
		if attrs[i] != "" {
			if err := p.Store(fmt.Sprintf("tables/%s/indexes/%s/%s/%s.json", table, i, attrs[i], attrs["id"]), attrs); err != nil {
				return "", err
			}
		}
	}

	return attrs["id"], nil
}
