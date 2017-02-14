package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/server"
)

func main() {
	// if err := server.New().Listen("tcp", ":9666"); err != nil {
	//   fmt.Printf("err = %+v\n", err)
	//   os.Exit(1)
	// }

	if err := server.New().Listen("unix", "/tmp/test.sock"); err != nil {
		fmt.Printf("err = %+v\n", err)
		os.Exit(1)
	}
}
