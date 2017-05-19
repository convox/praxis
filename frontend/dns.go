package frontend

import (
	"fmt"

	"github.com/miekg/dns"
)

type DNS struct {
	Host string

	frontend *Frontend
	mux      *dns.ServeMux
	server   *dns.Server
}

func NewDNS(host string, frontend *Frontend) *DNS {
	mux := dns.NewServeMux()

	mux.HandleFunc(".", resolvePassthrough)

	return &DNS{
		Host:     host,
		frontend: frontend,
		mux:      mux,
		server: &dns.Server{
			Addr:    fmt.Sprintf("%s:53", host),
			Handler: mux,
			Net:     "udp",
		},
	}
}

func (d *DNS) Serve() error {
	return d.server.ListenAndServe()
}

func (d *DNS) registerDomain(domain string) error {
	d.mux.HandleFunc(fmt.Sprintf("%s.", domain), d.resolveConvox)

	if err := d.setupResolver(domain); err != nil {
		return err
	}

	return nil
}

func (d *DNS) resolveConvox(w dns.ResponseWriter, r *dns.Msg) {
	log := Log.At("resolve.convox").Start()

	m := &dns.Msg{}
	m.SetReply(r)
	m.Compress = false
	m.RecursionAvailable = true
	m.Authoritative = true

	switch r.Opcode {
	case dns.OpcodeQuery:
		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				log = log.Append("name=%s", q.Name)
				if ip, ok := d.frontend.hosts[q.Name]; ok {
					if rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip)); err == nil {
						rr.Header().Ttl = 5
						m.Answer = append(m.Answer, rr)
						log.Successf("ip=%s", ip)
					}
				} else {
					log.Logf("error=%q", "unknown host")
				}
			}
		}
	}

	w.WriteMsg(m)
}

func resolvePassthrough(w dns.ResponseWriter, r *dns.Msg) {
	log := Log.At("resolve.passthrough").Start()

	c := dns.Client{Net: "tcp"}

	rs, _, err := c.Exchange(r, "8.8.8.8:53")
	if err != nil {
		m := &dns.Msg{}
		m.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(m)
		log.Error(err)
		return
	}

	w.WriteMsg(rs)

	for _, q := range r.Question {
		log.Successf("name=%q", q.Name)
	}
}
