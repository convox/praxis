package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/server"
)

func main() {
	if err := server.New().Listen("tcp", ":3000"); err != nil {
		fmt.Printf("err = %+v\n", err)
		os.Exit(1)
	}
}
