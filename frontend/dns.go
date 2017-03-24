package frontend

import (
	"fmt"

	"github.com/miekg/dns"
)

func startDns(ip string) error {
	dns.HandleFunc("convox.", resolveConvox)
	dns.HandleFunc(".", resolvePassthrough)

	server := &dns.Server{Addr: fmt.Sprintf("%s:53", ip), Net: "udp"}

	return server.ListenAndServe()
}

func resolveConvox(w dns.ResponseWriter, r *dns.Msg) {
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
				log = log.Namespace("name=%s", q.Name)
				if ip, ok := hosts[q.Name]; ok {
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
