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

	go func() {
		<-terminate
		if err := p.shutdown(); err != nil {
			fmt.Printf("ns=provider.local at=shutdown error=%q\n", err)
		}
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)

			if err := p.workerConverge(); err != nil {
				fmt.Printf("ns=provider.local at=converge error=%q\n", err)
			}
		}
	}()
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
