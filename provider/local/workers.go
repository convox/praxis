package local

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (p *Provider) Workers() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		p.shutdown()
	}()

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
			fmt.Fprintf(os.Stderr, "error converging %s: %s\n", a.Name, err)
			continue
		}
	}

	if err := p.convergePrune(); err != nil {
		return err
	}

	return nil
}
