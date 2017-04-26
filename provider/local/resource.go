package local

import (
	"fmt"
	"sort"
	"strings"

	"github.com/convox/praxis/types"
)

func (p *Provider) ResourceCreate(name, kind string, params map[string]string) (*types.Resource, error) {
	if _, err := p.ResourceGet(name); err == nil {
		return nil, fmt.Errorf("resource already exists: %s", name)
	}

	resource := &types.Resource{
		Name:   name,
		Status: "running",
		Type:   kind,
	}

	if err := p.storageStore(fmt.Sprintf("resources/%s/resource.json", resource.Name), resource); err != nil {
		return nil, err
	}

	return resource, nil
}

func (p *Provider) ResourceGet(name string) (*types.Resource, error) {
	var resource types.Resource

	if err := p.storageLoad(fmt.Sprintf("resources/%s/resource.json", name), &resource); err != nil {
		if strings.HasPrefix(err.Error(), "no such key:") {
			return nil, fmt.Errorf("no such resource: %s", name)
		}
		return nil, err
	}

	return &resource, nil
}

func (p *Provider) ResourceList() (types.Resources, error) {
	names, err := p.storageList("resources")
	if err != nil {
		return nil, err
	}

	resources := make(types.Resources, len(names))

	for i, name := range names {
		res, err := p.ResourceGet(name)
		if err != nil {
			return nil, err
		}

		resources[i] = *res
	}

	sort.Slice(resources, func(i, j int) bool { return resources[i].Name < resources[j].Name })

	return resources, nil
}
