package manifest

type Service struct {
	Name string `yaml:"-"`

	Build string `yaml:"build,omitempty"`
	Image string `yaml:"image,omitempty"`
}

type Services []Service
