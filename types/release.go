package types

import (
	"time"
)

type Release struct {
	Id string `json:"id"`

	App    string      `json:"app"`
	Build  string      `json:"build"`
	Env    Environment `json:"env"`
	Stage  int         `json:"stage"`
	Status string      `json:"status"`

	Created time.Time `json:"created"`
}

type Releases []Release

type ReleaseCreateOptions struct {
	Build string
	Env   map[string]string
	Stage int
}

type ReleaseListOptions struct {
	Count int
}
