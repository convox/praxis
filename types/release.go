package types

import "time"

type Release struct {
	Id string `json:"id"`

	App   string            `json:"app"`
	Build string            `json:"build"`
	Env   map[string]string `json:"env"`

	Created time.Time `json:"created"`
}

type Releases []Release

type ReleaseCreateOptions struct {
	Build string
	Env   map[string]string
}

func (v Releases) Len() int           { return len(v) }
func (v Releases) Less(i, j int) bool { return v[j].Created.Before(v[i].Created) }
func (v Releases) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
