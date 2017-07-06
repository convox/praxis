package main

import (
	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/types"
)

func main() {
	Rack, err := rack.NewFromEnv()
	if err != nil {
		panic(err)
	}

	Rack.ProcessRun("qa-app", types.ProcessRunOptions{})
}
