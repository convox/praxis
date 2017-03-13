package types

type Table struct {
	Name string

	Indexes []string
}

type Tables []Table

type TableCreateOptions struct {
	Indexes []string
}

type TableRow map[string]string

type TableRows []TableRow

type TableRowDeleteOptions struct {
	Index string
}

type TableRowGetOptions struct {
	Index string
}
