package local

import (
	"fmt"
	"os"
	"time"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
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
	log := Logger.At("converge").Start()

	apps, err := p.AppList()
	if err != nil {
		log.Error(err)
		return err
	}

	for _, a := range apps {
		if a.Release == "" {
			continue
		}

		r, err := p.ReleaseGet(a.Name, a.Release)
		if err != nil {
			log.Error(err)
			return err
		}

		b, err := p.BuildGet(a.Name, r.Build)
		if err != nil {
			log.Error(err)
			return err
		}

		m, err := manifest.Load([]byte(b.Manifest))
		if err != nil {
			log.Error(err)
			return err
		}

		for _, b := range m.Balancers {
			if err := p.registerBalancerWithFrontend(a.Name, b); err != nil {
				log.Error(err)
				return err
			}
		}

		ps, err := p.ProcessList(a.Name, types.ProcessListOptions{})
		if err != nil {
			log.Error(err)
			return err
		}

		counts := map[string]int{}

		for _, p := range ps {
			counts[p.Service] += 1
		}

		for _, s := range m.Services {
			for i := counts[s.Name]; i < s.Scale.Min; i++ {
				log.Logf("app=%s service=%s release=%s action=scale count=1", a.Name, s.Name, r.Id)

				if err := p.startService(m, a.Name, s.Name, r.Id); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	}

	log.Success()
	return nil
}
