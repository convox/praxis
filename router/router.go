package router

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/convox/praxis/api"
)

type Endpoint struct {
	Host    string        `json:"host"`
	IP      net.IP        `json:"ip"`
	Proxies map[int]Proxy `json:"proxies"`

	router *Router
}

type Router struct {
	Domain    string
	Interface string
	Subnet    string

	ca        tls.Certificate
	dns       *DNS
	endpoints map[string]Endpoint
	lock      sync.Mutex
	ip        net.IP
	net       *net.IPNet
}

func New(domain, iface, subnet string) (*Router, error) {
	ip, net, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	r := &Router{
		Domain:    domain,
		Interface: iface,
		Subnet:    subnet,
		endpoints: map[string]Endpoint{},
		ip:        ip,
		net:       net,
	}

	ca, err := caCertificate()
	if err != nil {
		return nil, err
	}

	r.ca = ca

	d, err := r.NewDNS()
	if err != nil {
		return nil, err
	}

	go d.Serve()

	r.dns = d

	return r, nil
}

func (r *Router) Serve() error {
	destroyInterface(r.Interface)

	if err := createInterface(r.Interface, r.ip.String()); err != nil {
		return err
	}

	defer destroyInterface(r.Interface)

	// reserve one ip for router
	r.endpoints[fmt.Sprintf("router.%s", r.Domain)] = Endpoint{IP: r.ip}

	rh := fmt.Sprintf("rack.%s", r.Domain)

	ep, err := r.createEndpoint(rh)
	if err != nil {
		return err
	}

	if _, err := r.createProxy(rh, fmt.Sprintf("https://%s:443", ep.IP), "https://localhost:5443"); err != nil {
		return err
	}

	a := api.New("convox.router", fmt.Sprintf("router.%s", r.Domain))

	a.Route("GET", "/endpoints", r.EndpointList)
	a.Route("POST", "/endpoints/{host}", r.EndpointCreate)
	a.Route("DELETE", "/endpoints/{host}", r.EndpointDelete)

	a.Route("POST", "/endpoints/{host}/proxies/{port}", r.ProxyCreate)

	err = a.Listen("h2", fmt.Sprintf("%s:443", r.ip))
	return err
}

func (r *Router) createEndpoint(host string) (*Endpoint, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if ep, ok := r.endpoints[host]; ok {
		return &ep, nil
	}

	ip, err := r.nextIP()
	if err != nil {
		return nil, err
	}

	if err := createAlias(r.Interface, ip.String()); err != nil {
		return nil, err
	}

	e := Endpoint{
		Host:    host,
		IP:      ip,
		Proxies: map[int]Proxy{},
		router:  r,
	}

	r.endpoints[host] = e

	return &e, nil
}

func (r *Router) createProxy(host, listen, target string) (*Proxy, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	ep, ok := r.endpoints[host]
	if !ok {
		return nil, fmt.Errorf("no such endpoint: %s", host)
	}

	ul, err := url.Parse(listen)
	if err != nil {
		return nil, err
	}

	ut, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	pi, err := strconv.Atoi(ul.Port())
	if err != nil {
		return nil, err
	}

	if p, ok := r.endpoints[host].Proxies[pi]; ok {
		return &p, nil
	}

	p, err := ep.NewProxy(host, ul, ut)
	if err != nil {
		return nil, err
	}

	r.endpoints[host].Proxies[pi] = *p

	go p.Serve()

	return p, nil
}

func (r *Router) hasIP(ip net.IP) bool {
	for _, e := range r.endpoints {
		if e.IP.Equal(ip) {
			return true
		}
	}

	return false
}

func (r *Router) nextIP() (net.IP, error) {
	ip := make(net.IP, len(r.ip))
	copy(ip, r.ip)

	for {
		if !r.hasIP(ip) {
			break
		}

		ip = incrementIP(ip)
	}

	return ip, nil
}

func incrementIP(ip net.IP) net.IP {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}

	return ip
}

func execute(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}
