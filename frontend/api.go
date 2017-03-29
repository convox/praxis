package frontend

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

const (
	cleanupInterval = 30 * time.Second
	endpointTTL     = 2 * time.Minute
)

var (
	endpoints = map[string]map[int]Endpoint{}
	hosts     = map[string]string{}
	lock      sync.Mutex
)

type Endpoint struct {
	Host   string    `json:"host"`
	Ip     string    `json:"ip"`
	Port   int       `json:"port"`
	Target string    `json:"target"`
	Until  time.Time `json:"until"`

	listener net.Listener
}

func startApi(ip, iface, subnet string) error {
	r := mux.NewRouter()

	r.HandleFunc("/endpoints", listEndpoints).Methods("GET")
	r.HandleFunc("/endpoints/{host}", createEndpoint(iface, subnet)).Methods("POST")
	r.HandleFunc("/endpoints/{host}", deleteEndpoint).Methods("DELETE")

	go cleanupEndpoints()

	return http.ListenAndServe(fmt.Sprintf("%s:9477", ip), r)
}

func cleanupEndpoints() {
	tick := time.Tick(cleanupInterval)

	for range tick {
		log := Log.At("endpoint.cleanup")

		for host, pe := range endpoints {
			log = log.Namespace("host=%q", host)

			for port, e := range pe {
				log = log.Namespace("port=%d", port)

				if e.Until.Before(time.Now()) {
					e.listener.Close()

					delete(endpoints[host], port)

					if len(endpoints[host]) == 0 {
						delete(endpoints, host)
					}

					log.Success()
				}
			}
		}
	}
}

func createEndpoint(iface, subnet string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()

		log := Log.At("endpoint.create").Start()

		host := mux.Vars(r)["host"]

		port := r.FormValue("port")
		target := r.FormValue("target")

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

		ip, err := ipForHost(iface, subnet, fmt.Sprintf("%s.", host))
		if err != nil {
			http.Error(w, "invalid port", 500)
			log.Error(err)
			return
		}

		log = log.Namespace("ip=%s", ip)

		ep := Endpoint{
			Host:   host,
			Ip:     ip,
			Port:   pi,
			Target: target,
			Until:  time.Now().Add(endpointTTL).UTC(),
		}

		_, exists := endpoints[ip][pi]

		if !exists {
			ln, err := createProxy(ip, port, target)
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Error(err)
				return
			}

			ep.listener = ln
		}

		endpoints[ip][pi] = ep

		data, err := json.MarshalIndent(ep, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Error(err)
			return
		}

		w.Write(data)

		log.Success()
	}
}

func deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
}

func listEndpoints(w http.ResponseWriter, r *http.Request) {
	ep := []Endpoint{}

	for _, m := range endpoints {
		for _, e := range m {
			ep = append(ep, e)
		}
	}

	data, err := json.MarshalIndent(ep, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(data)
}

func ipForHost(iface, subnet, host string) (string, error) {
	if ip, ok := hosts[host]; ok {
		return ip, nil
	}

	return createHost(iface, subnet, host)
}
