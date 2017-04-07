package frontend

import (
	"time"
)

type Endpoint struct {
	Host   string    `json:"host"`
	Port   int       `json:"port"`
	Ip     string    `json:"ip"`
	Target string    `json:"target"`
	Until  time.Time `json:"until"`

	proxy *Proxy
}

type Endpoints []Endpoint

func (e *Endpoint) Cleanup() error {
	if e.proxy != nil {
		if err := e.proxy.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (e Endpoints) Less(i, j int) bool {
	if e[i].Host == e[j].Host {
		return e[i].Port < e[i].Port
	}
	return e[i].Host < e[j].Host
}
