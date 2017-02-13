package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/builder"
	"github.com/convox/praxis/manifest"
)

func die(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	os.Exit(1)
}

func main() {
	if err := build(); err != nil {
		die(err)
	}
}

func build() error {
	m, err := manifest.LoadFile("manifest/testdata/full.yml")
	if err != nil {
		return err
	}

	b, err := builder.New(m, &builder.Options{
		Namespace: "test",
	})
	if err != nil {
		return err
	}

	fmt.Printf("m = %+v\n", m)
	fmt.Printf("b = %+v\n", b)

	if err := b.Build(); err != nil {
		return err
	}

	return nil
}
