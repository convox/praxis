package manifest

type Timer struct {
	Name string

	Command  string
	Schedule string
	Service  string
}

type Timers []Timer

func (t Timer) GetName() string {
	return t.Name
}
