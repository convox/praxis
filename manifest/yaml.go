package manifest

import (
	"fmt"
	"reflect"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func (v Environment) MarshalYAML() (interface{}, error) {
	env := []string{}

	for _, e := range v {
		if e.Default == nil {
			env = append(env, e.Key)
		} else {
			env = append(env, fmt.Sprintf("%s=%s", e.Key, e.Default))
		}
	}

	return env, nil
}

func (v *Environment) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var env []string

	if err := unmarshal(&env); err != nil {
		return err
	}

	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)

		switch len(parts) {
		case 1:
			*v = append(*v, EnvironmentPair{
				Key:     parts[0],
				Default: nil,
			})
		case 2:
			*v = append(*v, EnvironmentPair{
				Key:     parts[0],
				Default: &parts[1],
			})
		default:
			return fmt.Errorf("could not parse environment: %s", e)
		}
	}

	return nil
}

func (v Services) MarshalYAML() (interface{}, error) {
	return marshalMap(v, "Name")
}

func (v *Services) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshalMap(unmarshal, v, "Name")
}

func (v Tables) MarshalYAML() (interface{}, error) {
	return marshalMap(v, "Name")
}

func (v *Tables) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshalMap(unmarshal, v, "Name")
}

func (v Volumes) MarshalYAML() (interface{}, error) {
	volumes := []string{}

	for _, vol := range v {
		volumes = append(volumes, fmt.Sprintf("%s:%s", vol.Local, vol.Remote))
	}

	return volumes, nil
}

func (v *Volumes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var volumes []string

	if err := unmarshal(&volumes); err != nil {
		return err
	}

	for _, vol := range volumes {
		parts := strings.SplitN(vol, ":", 2)

		switch len(parts) {
		case 1:
			*v = append(*v, Volume{
				Local:  parts[0],
				Remote: parts[0],
			})
		case 2:
			*v = append(*v, Volume{
				Local:  parts[0],
				Remote: parts[1],
			})
		default:
			return fmt.Errorf("could not parse volume: %s", vol)
		}
	}

	return nil
}

func marshalMap(v interface{}, key string) (interface{}, error) {
	ms := yaml.MapSlice{}

	vv := reflect.ValueOf(v)

	if vv.Kind() != reflect.Slice {
		return nil, fmt.Errorf("item is not a slice")
	}

	for i := 0; i < vv.Len(); i++ {
		vi := vv.Index(i)
		kf := vi.FieldByName(key)

		if kf.Kind() != reflect.String {
			return nil, fmt.Errorf("can not find key: %s", key)
		}

		ms = append(ms, yaml.MapItem{
			Key:   kf.Interface(),
			Value: vv.Index(i).Interface(),
		})
	}

	return ms, nil
}

func unmarshalMap(unmarshal func(interface{}) error, out interface{}, key string) error {
	var om yaml.MapSlice

	if err := unmarshal(&om); err != nil {
		return err
	}

	ov := reflect.ValueOf(out)

	if ov.Kind() != reflect.Ptr {
		return fmt.Errorf("could not unmarshal")
	}

	if reflect.Indirect(ov).Kind() != reflect.Slice {
		return fmt.Errorf("could not unmarshal")
	}

	ovi := ov.Elem()

	mt := reflect.Indirect(reflect.ValueOf(out)).Type().Elem()
	mpt := reflect.New(mt).Type()
	mm := reflect.New(reflect.MapOf(reflect.TypeOf(""), mpt))

	if err := unmarshal(mm.Interface()); err != nil {
		return err
	}

	for _, oi := range om {
		k, ok := oi.Key.(string)
		if !ok {
			return fmt.Errorf("unknown key type: %v", k)
		}

		mi := mm.Elem().MapIndex(reflect.ValueOf(k)).Elem()

		if !mi.IsValid() {
			mi = reflect.New(mt).Elem()
		}

		kf := mi.FieldByName(key)

		if !kf.CanSet() || kf.Kind() != reflect.String {
			return fmt.Errorf("can not set key: %s", key)
		}

		kf.SetString(k)

		if !mi.Type().ConvertibleTo(mt) {
			return fmt.Errorf("could not unmarshal")
		}

		ovi.Set(reflect.Append(ovi, mi))
	}

	return nil
}
