package main

import (
	"flag"
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

	var port = flag.Int("port", 3000, "port to listen on")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)

	if err := s.Listen("tcp", addr); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
