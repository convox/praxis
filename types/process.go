package types

import (
	"io"
	"time"
)

type Process struct {
	Id string `json:"id"`

	App     string    `json:"app"`
	Command string    `json:"command"`
	Release string    `json:"release"`
	Service string    `json:"service"`
	Started time.Time `json:"started"`
	Type    string    `json:"type"`
}

type Processes []Process

type ProcessExecOptions struct {
	Height int
	Width  int
}

type ProcessListOptions struct {
	Service string
	Type    string
}

type ProcessRunOptions struct {
	Command     string
	Environment map[string]string
	Height      int
	Image       string
	Links       []string
	Name        string
	Ports       map[int]int
	Release     string
	Service     string
	Stream      io.ReadWriter
	Type        string
	Volumes     map[string]string
	Width       int
}
