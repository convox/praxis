package types

import "io"

type System struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	Version string `json:"version"`
}

type SystemInstallOptions struct {
	Color   bool
	Key     string
	Output  io.Writer
	Version string
}
