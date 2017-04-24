package frontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

const (
	cleanupInterval = 30 * time.Second
	endpointTTL     = 30 * time.Minute
)

type API struct {
	Host string

	frontend *Frontend
	lock     sync.Mutex
	router   *mux.Router
}

func NewAPI(host string, frontend *Frontend) *API {
	r := mux.NewRouter()

	api := &API{
		Host:     host,
		frontend: frontend,
		router:   r,
	}

	r.HandleFunc("/endpoints", api.listEndpoints).Methods("GET")
	r.HandleFunc("/endpoints/{host}", api.createEndpoint).Methods("POST")
	r.HandleFunc("/endpoints/{host}", api.deleteEndpoint).Methods("DELETE")

	go api.Cleanup()

	return api
}

func (a *API) Serve() error {
	return http.ListenAndServe(fmt.Sprintf("%s:9477", a.Host), a.router)
}

func (a *API) Cleanup() {
	log := Log.At("endpoint.cleanup")
	tick := time.Tick(cleanupInterval)

	for range tick {
		for hash, e := range a.frontend.endpoints {
			log := log.Namespace("host=%q port=%d", e.Host, e.Port)

			if e.Until.Before(time.Now()) {
				if err := e.Cleanup(); err != nil {
					log.Error(err)
					continue
				}

				delete(a.frontend.endpoints, hash)
				log.Success()
			}
		}
	}
}

func (a *API) createEndpoint(w http.ResponseWriter, r *http.Request) {
	a.lock.Lock()
	defer a.lock.Unlock()

	log := Log.At("endpoint.create").Start()

	host := mux.Vars(r)["host"]
	port := r.FormValue("port")
	target := r.FormValue("target")
	parts := strings.Split(host, ".")
	domain := parts[len(parts)-1]

	defer r.Body.Close()

	if host == "" {
		http.Error(w, "host required", 500)
		log.Logf("error=%q", "host required")
		return
	}

	if port == "" {
		http.Error(w, "port required", 500)
		log.Logf("error=%q", "port required")
		return
	}

	if target == "" {
		http.Error(w, "target required", 500)
		log.Logf("error=%q", "target required")
		return
	}

	log = log.Namespace("host=%s port=%s target=%s", host, port, target)

	pi, err := strconv.Atoi(port)
	if err != nil {
		http.Error(w, "invalid port", 500)
		log.Error(err)
		return
	}

	ip, err := a.ipForHost(a.frontend.Interface, a.frontend.Subnet, fmt.Sprintf("%s.", host))
	if err != nil {
		http.Error(w, "invalid port", 500)
		log.Error(err)
		return
	}

	log = log.Namespace("ip=%s", ip)

	hash := fmt.Sprintf("%s:%d", ip, pi)

	ep, ok := a.frontend.endpoints[hash]
	if !ok {
		proxy := NewProxy(ip, port, target, a.frontend)

		go proxy.Serve()

		ep = Endpoint{
			Host:   host,
			Ip:     ip,
			Port:   pi,
			Target: target,
			proxy:  proxy,
		}
	}

	ep.Target = target
	ep.Until = time.Now().Add(endpointTTL).UTC()

	a.frontend.endpoints[hash] = ep

	if _, exists := a.frontend.domains[domain]; !exists {
		if err := a.frontend.DNS.registerDomain(domain); err != nil {
			http.Error(w, err.Error(), 500)
			log.Error(err)
			return
		}

		a.frontend.domains[domain] = true
	}

	data, err := json.MarshalIndent(ep, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Error(err)
		return
	}

	w.Write(data)

	log.Success()
}

func (a *API) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	a.lock.Lock()
	defer a.lock.Unlock()
}

func (a *API) listEndpoints(w http.ResponseWriter, r *http.Request) {
	ep := Endpoints{}

	for _, e := range a.frontend.endpoints {
		ep = append(ep, e)
	}

	sort.Slice(ep, ep.Less)

	data, err := json.MarshalIndent(ep, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(data)
}

func (a *API) ipForHost(iface, subnet, host string) (string, error) {
	if ip, ok := a.frontend.hosts[host]; ok {
		return ip, nil
	}

	return a.frontend.createHost(host)
}
