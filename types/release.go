package types

import "time"

type Release struct {
	Id string `json:"id"`

	App   string            `json:"app"`
	Build string            `json:"build"`
	Env   map[string]string `json:"env"`

	Created time.Time `json:"created"`
}

type ReleaseCreateOptions struct {
	Build string
	Env   map[string]string
}
