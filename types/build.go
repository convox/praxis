package types

import "time"

type Build struct {
	Id       string `json:"id"`
	App      string `json:"app"`
	Manifest string `json:"manifest"`
	Release  string `json:"release"`
	Status   string `json:"status"`

	Started time.Time `json:"started"`
	Ended   time.Time `json:"ended"`
}

type BuildCreateOptions struct {
	Cache    bool
	Manifest string
}
