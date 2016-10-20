package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/api"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}

func start() error {
	return api.Listen("0.0.0.0:9877")
}
