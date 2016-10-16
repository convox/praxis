package main

import (
	"os"

	"github.com/convox/praxis/cli"
	"github.com/convox/praxis/client"
)

func main() {
	cli.Init(cli.Settings{
		Name:    "cx",
		Summary: "build, run, and deploy convox apps",
		Description: `
			convox is a command

			line
			  application
		`,
	})

	cli.Run(os.Args)
}

func rack() *client.Client {
	return client.New("http://localhost:9877/apps/system")
}
