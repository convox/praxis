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
			if err := p.workerConverge(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
			}
		}
	}
}

func (p *Provider) workerConverge() error {
	apps, err := p.AppList()
	if err != nil {
		return err
	}

	for _, a := range apps {
		if err := p.converge(a.Name); err != nil {
			return err
		}
	}

	return nil
}
