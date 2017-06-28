package router

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/types"

	mrand "math/rand"
)

type Proxy struct {
	Listen *url.URL
	Target *url.URL

	endpoint *Endpoint
}

func (e *Endpoint) NewProxy(host string, listen, target *url.URL) (*Proxy, error) {
	p := &Proxy{
		Listen:   listen,
		Target:   target,
		endpoint: e,
	}

	return p, nil
}

func (p Proxy) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"listen": p.Listen.String(),
		"target": p.Target.String(),
	})
}

func (p *Proxy) Serve() error {
	ln, err := net.Listen("tcp", p.Listen.Host)
	if err != nil {
		return err
	}

	defer ln.Close()

	switch p.Listen.Scheme {
	case "https", "tls":
		cert, err := p.endpoint.router.generateCertificate(p.endpoint.Host)
		if err != nil {
			return err
		}

		cfg := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		// TODO: check for h2
		cfg.NextProtos = []string{"h2"}

		ln = tls.NewListener(ln, cfg)
	}

	switch p.Listen.Scheme {
	case "http", "https":
		if err := http.Serve(ln, proxyHTTP(p.Listen, p.Target)); err != nil {
			return err
		}
	case "tcp":
		if err := proxyTCP(ln, p.Target); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown listener scheme: %s", p.Listen.Scheme)
	}

	return nil
}

func proxyHTTP(listen, target *url.URL) http.Handler {
	if target.Hostname() == "rack" {
		return http.HandlerFunc(proxyRackHTTP(listen, target))
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	tr := logTransport{
		Transport: http.DefaultTransport.(*http.Transport),
		listener:  listen,
	}

	// tr.DialContext = proxyDialer(target)

	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	proxy.Transport = tr

	return proxy
}

func proxyTCP(listener net.Listener, target *url.URL) error {
	for {
		cn, err := listener.Accept()
		if err != nil {
			return err
		}

		go proxyRackTCP(cn, target)
	}
}

func proxyTCPConnection(cn net.Conn, target *url.URL) error {
	if target.Hostname() == "rack" {
		return proxyRackTCP(cn, target)
	}

	defer cn.Close()

	fmt.Printf("target = %+v\n", target)

	oc, err := net.Dial("tcp", target.Host)
	if err != nil {
		return err
	}

	defer oc.Close()

	return helpers.Pipe(cn, oc)
}

func proxyRackTCP(cn net.Conn, target *url.URL) error {
	defer cn.Close()

	parts := strings.Split(target.Path, "/")

	if len(parts) < 4 {
		return fmt.Errorf("invalid rack endpoint: %s", target)
	}

	app := parts[1]
	kind := parts[2]
	rp := strings.Split(parts[3], ":")

	if len(rp) < 2 {
		return fmt.Errorf("invalid %s endpoint: %s", kind, parts[2])
	}

	resource := rp[0]

	var pr io.ReadCloser

	r, err := rack.NewFromEnv()
	if err != nil {
		return err
	}

	switch kind {
	case "resource":
		rc, err := r.ResourceProxy(app, resource, cn)
		if err != nil {
			return err
		}
		pr = rc
	default:
		return fmt.Errorf("unknown proxy type: %s", kind)
	}

	if _, err := io.Copy(cn, pr); err != nil {
		return err
	}

	return nil
}

func proxyRackHTTP(listen, target *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(target.Path, "/")

		if len(parts) < 4 {
			http.Error(w, "invalid rack endpoint", 500)
			return
		}

		app := parts[1]
		kind := parts[2]
		sp := strings.Split(parts[3], ":")

		if len(sp) < 2 {
			http.Error(w, fmt.Sprintf("invalid %s endpoint: %s", kind, parts[2]), 500)
			return
		}

		service := sp[0]

		p, err := strconv.Atoi(sp[1])
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		port := p

		c := &http.Client{}

		switch kind {
		case "service":
			fmt.Printf("ns=convox.router at=proxy type=http listen=%q target=rack app=%q service=%q port=%d path=%q\n", listen, app, service, port, r.URL.Path)
			c.Transport = serviceTransport(app, service, port)
		default:
			http.Error(w, fmt.Sprintf("unknown proxy type: %s", kind), 500)
			return
		}

		rurl := fmt.Sprintf("%s://%s%s", target.Scheme, r.Host, r.URL.Path)

		creq, err := http.NewRequest(r.Method, rurl, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		for h, vs := range r.Header {
			for _, v := range vs {
				creq.Header.Add(h, v)
			}
		}

		creq.Header.Add("X-Forwarded-For", r.RemoteAddr)
		creq.Header.Add("X-Forwarded-Port", listen.Port())
		creq.Header.Add("X-Forwarded-Proto", listen.Scheme)

		res, err := c.Do(creq)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer res.Body.Close()

		w.WriteHeader(res.StatusCode)

		for k, vs := range res.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}

		if _, err := io.Copy(w, res.Body); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

func serviceTransport(app, service string, port int) http.RoundTripper {
	tr := http.DefaultTransport.(*http.Transport)

	tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		r, err := rack.NewFromEnv()
		if err != nil {
			return nil, err
		}

		pss, err := r.ProcessList(app, types.ProcessListOptions{Service: service})
		if err != nil {
			return nil, err
		}

		if len(pss) < 1 {
			return nil, fmt.Errorf("no processes available for service: %s", service)
		}

		ps := pss[mrand.Intn(len(pss))]

		a, b := net.Pipe()

		go func() {
			pr, err := r.ProcessProxy(app, ps.Id, port, a)
			if err != nil {
				return
			}

			io.Copy(a, pr)
		}()

		return b, nil
	}

	return tr
}
