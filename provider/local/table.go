package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

const (
	SortableTime = "20060102.150405.000000000"
)

func (p *Provider) TableLoad(app, table, id string) (map[string]string, error) {
	r, err := p.load(fmt.Sprintf("/tables/%s/%s/index/id/%s", app, table, id))
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var attrs map[string]string

	if err := json.Unmarshal(data, &attrs); err != nil {
		return nil, err
	}

	return attrs, nil
}

func (p *Provider) TableRemove(app, table, id string) error {
	return p.remove(fmt.Sprintf("/tables/%s/%s/index/id/%s", app, table, id))
}

func (p *Provider) TableSave(app, table, id string, attrs map[string]string) error {
	if attrs == nil {
		attrs = map[string]string{}
	}

	if _, ok := attrs["created"]; !ok {
		attrs["created"] = time.Now().Format(SortableTime)
	}

	attrs["updated"] = time.Now().Format(SortableTime)

	data, err := json.Marshal(attrs)
	if err != nil {
		return err
	}

	if err := p.save(fmt.Sprintf("/tables/%s/%s/index/id/%s", app, table, id), bytes.NewReader(data)); err != nil {
		return err
	}

	return nil
}
