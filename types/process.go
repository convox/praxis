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
	Status  string    `json:"status"`
	Type    string    `json:"type"`
}

type Processes []Process

type ProcessExecOptions struct {
	Height int
	Width  int

	Input  io.Reader
	Output io.Writer
}

type ProcessListOptions struct {
	Service string
}

type ProcessRunOptions struct {
	Command     string
	Environment map[string]string
	Height      int
	Image       string
	Links       []string
	Memory      int
	Name        string
	Ports       map[int]int
	Release     string
	Service     string
	Volumes     map[string]string
	Width       int

	Input  io.Reader
	Output io.Writer
}
