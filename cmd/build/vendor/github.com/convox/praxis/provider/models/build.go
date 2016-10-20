package models

import "time"

type Build struct {
	Id string `json:"id"`

	Error    string    `json:"error,omitempty"`
	Ended    time.Time `json:"ended"`
	Logs     string    `json:"logs"`
	Manifest string    `json:"manifest"`
	Process  string    `json:"process"`
	Status   string    `json:"status"`
}

type BuildCreateOptions struct {
	Cache bool
}
