package manifest

import (
	"fmt"
	"reflect"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func (v Balancers) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (v *Balancers) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshalMapSlice(unmarshal, v)
}

func (v *Balancer) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshalMapSlice(unmarshal, &v.Endpoints)
}

func (v *Balancer) SetName(name string) {
	v.Name = name
}

func (v BalancerEndpoints) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

// func (v *BalancerEndpoints) UnmarshalYAML(unmarshal func(interface{}) error) error {
//   return unmarshalMapSlice(unmarshal, v)
// }

func (v *BalancerEndpoint) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var w interface{}

	if err := unmarshal(&w); err != nil {
		return err
	}

	switch t := w.(type) {
	case string:
		v.Target = t
	default:
		return fmt.Errorf("unknown type for endpoint: %T", t)
	}

	return nil
}

func (v *BalancerEndpoint) SetName(name string) {
	parts := strings.Split(name, "/")

	switch len(parts) {
	case 1:
		v.Port = parts[0]
	case 2:
		v.Port = parts[0]
		v.Protocol = parts[1]
	case 3:
		v.Port = parts[0]
		v.Protocol = parts[1]

		switch parts[2] {
		case "301":
			v.Redirect = v.Target
			v.Target = ""
		}
	}
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

func (v *ServiceBuild) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var w interface{}

	if err := unmarshal(&w); err != nil {
		return err
	}

	switch t := w.(type) {
	case map[interface{}]interface{}:
		type serviceBuild ServiceBuild
		var r serviceBuild
		if err := remarshal(w, &r); err != nil {
			return err
		}
		v.Args = r.Args
		v.Path = r.Path
	case string:
		v.Path = t
	default:
		return fmt.Errorf("unknown type for service build: %T", t)
	}

	return nil
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

		if err := remarshal(msi.Value, item); err != nil {
			return err
		}

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

		rv.Set(reflect.Append(rv, reflect.ValueOf(item).Elem()))
	}

	return nil
}
