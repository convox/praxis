package types

import "io"

type InstallOptions struct {
	Color   bool
	Output  io.Writer
	Version string
}
