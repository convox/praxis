package types

import "time"

type Build struct {
	Id       string `json:"id"`
	App      string `json:"app"`
	Manifest string `json:"manifest"`
	Process  string `json:"process"`
	Release  string `json:"release"`
	Status   string `json:"status"`

	Created time.Time `json:"created"`
	Started time.Time `json:"started"`
	Ended   time.Time `json:"ended"`
}

type Builds []Build

type BuildCreateOptions struct {
	Cache    bool
	Manifest string
	Stage    int
}

type BuildUpdateOptions struct {
	Ended    time.Time
	Manifest string
	Release  string
	Started  time.Time
	Status   string
}
