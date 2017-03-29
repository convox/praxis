package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/server"
)

func main() {
	s := server.New()

	if err := s.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	if err := s.Listen("tcp", ":3000"); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
