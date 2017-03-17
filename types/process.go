package types

import (
	"io"
	"time"
)

type Process struct {
	Id string `json:"id"`

	App     string `json:"app"`
	Command string `json:"command"`
	Release string `json:"release"`
	Service string `json:"service"`

	Started time.Time `json:"started"`
}

type Processes []Process

type ProcessExecOptions struct {
	Height int
	Width  int
}

type ProcessListOptions struct {
	Service string
}

type ProcessRunOptions struct {
	Command     string
	Environment map[string]string
	Height      int
	Name        string
	Release     string
	Service     string
	Stream      io.ReadWriter
	Volumes     map[string]string
	Width       int
}

type ProcessStartOptions struct {
	Command     string
	Environment map[string]string
	Image       string
	Name        string
	Release     string
	Service     string
	Volumes     map[string]string
}
