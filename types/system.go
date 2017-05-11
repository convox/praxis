package types

import "io"

type System struct {
	Account string `json:"account"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Region  string `json:"region"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

type SystemInstallOptions struct {
	Color    bool
	Output   io.Writer
	Password string
	Version  string
}

type SystemUpdateOptions struct {
	Output   io.Writer
	Password string
	Version  string
}
