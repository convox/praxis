package manifest

type Queue struct {
	Name string `yaml:"-"`

	Timeout int `yaml:"timeout,omitempty"`
}

type Queues []Queue
