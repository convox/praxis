package main

import (
	"log"

	"github.com/convox/praxis/frontend"
)

const (
	iname  = "vlan0"
	subnet = "10.42.84"
)

func main() {
	if err := frontend.Serve(iname, subnet); err != nil {
		log.Fatal(err)
	}
}
