package types

type Table struct {
	Name string

	Indexes []string
}

type Tables []Table

func (v Tables) Less(i, j int) bool { return v[i].Name < v[j].Name }

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
