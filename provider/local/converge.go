package local

import (
	"fmt"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func (p *Provider) converge() error {
	log := Logger.At("converge").Start()

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("r = %+v\n", r)
		}
	}()

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

		rlog := fmt.Sprintf("apps/%s/releases/%s/log", a.Name, r.Id)

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

		sc, err := p.serviceCounts(a.Name)
		if err != nil {
			log.Error(err)
			return err
		}

		for _, b := range m.Balancers {
			if !p.balancerRunning(a.Name, b) {
				log.Logf("app=%s balancer=%s action=start", a.Name, b.Name)
				p.storageLogWrite(rlog, []byte(fmt.Sprintf("starting balancer: %s\n", b.Name)))

				if err := p.balancerStart(a.Name, b); err != nil {
					log.Error(err)
					return err
				}
			}

			if err := p.balancerRegister(a.Name, b); err != nil {
				log.Error(err)
				return err
			}
		}

		for _, s := range m.Services {
			for i := sc[s.Name]; i < s.Scale.Min; i++ {
				log.Logf("app=%s service=%s release=%s action=start", a.Name, s.Name, r.Id)
				p.storageLogWrite(rlog, []byte(fmt.Sprintf("starting service: %s\n", s.Name)))

				if err := p.serviceStart(m, a.Name, s.Name, r.Id); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	}

	log.Success()
	return nil
}

func (p *Provider) serviceCounts(app string) (map[string]int, error) {
	ps, err := p.ProcessList(app, types.ProcessListOptions{})
	if err != nil {
		return nil, err
	}

	counts := map[string]int{}

	for _, p := range ps {
		counts[p.Service] += 1
	}

	return counts, nil
}
