package manifest

type Timer struct {
	Name string

	Command  string
	Schedule string
	Service  string
}

type Timers []Timer
