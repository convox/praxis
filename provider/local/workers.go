package local

import (
	"fmt"
	"os"
	"time"
)

func (p *Provider) workers() {
	converge := time.Tick(5 * time.Second)

	for {
		select {
		case <-converge:
			if err := p.converge(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
			}
		}
	}
}
