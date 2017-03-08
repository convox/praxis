package types

type Table struct {
	Name string

	Indexes []string
}

type Tables []Table

type TableFetchOptions struct {
	Index string
}
