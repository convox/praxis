package main

import (
	"fmt"
	"os"

	"github.com/convox/praxis/fsync/watch"
)

func main() {
	ch := make(chan watch.Change)

	for _, w := range os.Args[1:] {
		go watch.Watch(w, ch)
	}

	for c := range ch {
		fmt.Printf("%s|%s|%s\n", c.Operation, c.Base, c.Path)
	}
}
