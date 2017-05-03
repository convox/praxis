package types

import (
	"time"
)

type Release struct {
	Id string `json:"id"`

	App    string            `json:"app"`
	Build  string            `json:"build"`
	Error  string            `json:"error"`
	Env    map[string]string `json:"env"`
	Status string            `json:"status"`

	Created time.Time `json:"created"`
}

type Releases []Release

type ReleaseCreateOptions struct {
	Build string
	Env   map[string]string
}

type ReleaseListOptions struct {
	Count int
}
