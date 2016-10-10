package manifest

type Table struct {
	Name string `yaml:"-"`

	Indexes []string `yaml:"indexes,omitempty"`
}

type Tables []Table
