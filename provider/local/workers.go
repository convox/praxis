package local

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (p *Provider) Workers() {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	converge := time.Tick(5 * time.Second)

	for {
		select {
		case <-terminate:
			if err := p.shutdown(); err != nil {
				fmt.Printf("ns=provider.local at=shutdown error=%q\n", err)
			}
		case <-converge:
			if err := p.workerConverge(); err != nil {
				fmt.Printf("ns=provider.local at=converge error=%q\n", err)
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
			fmt.Printf("ns=provider.local at=converge app=%s error=%q\n", a.Name, err)
			continue
		}
	}

	if err := p.convergePrune(); err != nil {
		return err
	}

	return nil
}
