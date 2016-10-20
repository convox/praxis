package models

type Process struct {
	Id string `json:"id"`
}

type ProcessRunOptions struct {
	Command     []string
	Environment map[string]string
}
