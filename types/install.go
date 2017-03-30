package types

import "io"

type InstallOptions struct {
	Color   bool
	Key     string
	Output  io.Writer
	Version string
}
