package provider

import "time"

type App struct {
	Name string `json:"name"`
}

type Apps []App

type Build struct {
	Id string `json:"id"`

	Error    string    `json:"error,omitempty"`
	Ended    time.Time `json:"ended"`
	Logs     string    `json:"logs"`
	Manifest string    `json:"manifest"`
	Process  string    `json:"process"`
	Release  string    `json:"release"`
	Status   string    `json:"status"`
}

type Builds []Build

type Environment map[string]string

type Process struct {
	Id string `json:"id"`
}

type Release struct {
	Id string `json:"id"`

	Build       string      `json:"build"`
	Environment Environment `json:"environment"`
}
