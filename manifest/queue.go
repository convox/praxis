package manifest

type Queue struct {
	Name string

	Timeout string
}

type Queues []Queue
