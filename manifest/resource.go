package manifest

type Resource struct {
	Name  string `yaml:"-"`
	Type  string `yaml:"type"`
	Port  string `yaml:"port,omitempty"`
	Image string `yaml:"image,omitempty"`
}

type Resources []Resource

func (r Resource) GetName() string {
	return r.Name
}
