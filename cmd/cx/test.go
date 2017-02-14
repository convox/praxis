package main

import (
	"bytes"
	"fmt"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "test",
		Description: "run tests",
		Action:      runTest,
	})
}

func runTest(c *cli.Context) error {
	// app, err := Rack.AppCreate("")
	// if err != nil {
	//   return err
	// }

	// release, err := releaseDirectory(app.Name, ".")
	// if err != nil {
	//   return err
	// }

	// data, err := Rack.ReleaseManifest(app.Name, release.Id)
	// if err != nil {
	//   return err
	// }

	// m, err := manifest.Load(data)
	// if err != nil {
	//   return err
	// }

	// for _, s := range m.Services {
	//   if s.Test != "" {
	//     _, err := Rack.ProcessRun(app.Name, s.Name, rack.ProcessRunOptions{
	//       Output: os.Stdout,
	//     })
	//     if err != nil {
	//       return err
	//     }
	//   }
	// }

	// fmt.Printf("m = %+v\n", m)

	// r, err := Rack.GetStream("/clockstream")
	// if err != nil {
	//   return err
	// }

	// io.Copy(os.Stdout, r)

	fmt.Println("success")

	return nil
}

func releaseDirectory(app, dir string) (*rack.Release, error) {
	context := bytes.NewReader([]byte{})

	object, err := Rack.ObjectStore("", context)
	if err != nil {
		return nil, err
	}

	build, err := Rack.BuildCreate(app, object.URL)
	if err != nil {
		return nil, err
	}

	return Rack.ReleaseCreate(app, rack.ReleaseCreateOptions{
		Build: build.Id,
	})
}
