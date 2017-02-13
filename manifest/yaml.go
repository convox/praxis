package manifest

import (
	"fmt"
	"reflect"

	yaml "gopkg.in/yaml.v2"
)

func (v Balancers) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (v *Balancers) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshalMapSlice(unmarshal, v)
}

func (v *Balancer) SetName(name string) {
	v.Name = name
}

func (v BalancerEndpoints) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (v *BalancerEndpoints) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshalMapSlice(unmarshal, v)
}

func (v *BalancerEndpoint) SetName(name string) {
	v.Port = name
}

func (v Queues) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (v *Queues) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshalMapSlice(unmarshal, v)
}

func (v *Queue) SetName(name string) {
	v.Name = name
}

func (v Services) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (v *Services) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshalMapSlice(unmarshal, v)
}

func (v *Service) SetName(name string) {
	v.Name = name
}

func (v Tables) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (v *Tables) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshalMapSlice(unmarshal, v)
}

func (v *Table) SetName(name string) {
	v.Name = name
}

type NameSetter interface {
	SetName(name string)
}

func remarshal(in, out interface{}) error {
	data, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, out)
}

func unmarshalMapSlice(unmarshal func(interface{}) error, v interface{}) error {
	rv := reflect.ValueOf(v).Elem()
	vit := rv.Type().Elem()

	var ms yaml.MapSlice

	if err := unmarshal(&ms); err != nil {
		return err
	}

	for _, msi := range ms {
		item := reflect.New(vit).Interface()

		if ns, ok := item.(NameSetter); ok {
			switch t := msi.Key.(type) {
			case int:
				ns.SetName(fmt.Sprintf("%d", t))
			case string:
				ns.SetName(t)
			default:
				return fmt.Errorf("unknown key type: %T", t)
			}
		}

		if err := remarshal(msi.Value, item); err != nil {
			return err
		}

		rv.Set(reflect.Append(rv, reflect.ValueOf(item).Elem()))
	}

	return nil
}
