package types

import (
	"io"
	"time"
)

// Process represents a running Process
type Process struct {
	Id string `json:"id"`

	App     string `json:"app"`
	Release string `json:"release"`
	Service string `json:"service"`

	Started time.Time `json:"started"`
}

// ProcessExecOptions are options for ProcessExec
type ProcessExecOptions struct {
	Height int
	Width  int
}

// ProcessRunOptions are options for ProcessRun
type ProcessRunOptions struct {
	Command string
	Height  int
	Width   int
	Release string
	Stream  io.ReadWriter
}
