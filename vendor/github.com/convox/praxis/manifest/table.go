package manifest

type Table struct {
	Name string

	Indexes []string
}

type Tables []Table
