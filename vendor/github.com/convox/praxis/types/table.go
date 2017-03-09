package types

type Table struct {
	Name string

	Indexes []string
}

type Tables []Table

type TableCreateOptions struct {
	Indexes []string
}

type TableFetchOptions struct {
	Index string
}
