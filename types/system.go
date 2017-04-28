package types

import "io"

type System struct {
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Version string `json:"version"`
}

type SystemInstallOptions struct {
	Color    bool
	Output   io.Writer
	Password string
	Version  string
}

type SystemUpdateOptions struct {
	Output  io.Writer
	Version string
}
