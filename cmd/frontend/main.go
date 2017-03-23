package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/gorilla/mux"
	"github.com/miekg/dns"
)

const (
	subnet = "10.42.84"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}
}

func start() error {
	ip, err := setupListener()
	if err != nil {
		return err
	}

	go startDns(ip)
	go startApi(ip)

	select {}

	return nil
}

func startApi(ip string) error {
	r := mux.NewRouter()

	r.HandleFunc("/endpoints", listEndpoints).Methods("GET")
	r.HandleFunc("/endpoints", createEndpoint).Methods("POST")
	r.HandleFunc("/endpoints/{ip}", deleteEndpoint).Methods("DELETE")

	return http.ListenAndServe(fmt.Sprintf("%s:9477", ip), r)
}

func startDns(ip string) error {
	dns.HandleFunc("convox.", resolveConvox)
	dns.HandleFunc(".", resolvePassthrough)

	server := &dns.Server{Addr: fmt.Sprintf("%s:53", ip), Net: "udp"}

	return server.ListenAndServe()
}

func createListener(name string) (string, error) {
	cmd := exec.Command("ifconfig", name, "create")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil
	}

	ip := fmt.Sprintf("%s.0", subnet)

	cmd = exec.Command("ifconfig", name, ip, "netmask", "255.255.255.255", "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil
	}

	return ip, nil
}

func destroyListener(name string) error {
	return exec.Command("ifconfig", name, "destroy").Run()
}

func setupListener() (string, error) {
	destroyListener("vlan0")
	return createListener("vlan0")
}

func resolveConvox(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(r)
	m.Compress = false
	m.RecursionAvailable = true

	switch r.Opcode {
	case dns.OpcodeQuery:
		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				if ip, ok := hosts[q.Name]; ok {
					if rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip)); err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}
			}
		}
	}

	w.WriteMsg(m)
}

func resolvePassthrough(w dns.ResponseWriter, r *dns.Msg) {
	c := dns.Client{Net: "tcp"}

	rs, _, err := c.Exchange(r, "8.8.8.8:53")
	if err != nil {
		m := &dns.Msg{}
		m.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(m)
		return
	}

	w.WriteMsg(rs)
}

var hosts = map[string]string{}
var endpoints = map[string]string{}
var lock sync.Mutex

func createEndpoint(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()

	addr := r.FormValue("addr")
	host := r.FormValue("host")
	port := r.FormValue("port")

	if addr == "" {
		http.Error(w, "addr required", 500)
		return
	}

	if host == "" {
		http.Error(w, "host required", 500)
		return
	}

	if port == "" {
		http.Error(w, "port required", 500)
		return
	}

	ip := fmt.Sprintf("%s.%d", subnet, len(endpoints)+1)

	hosts[fmt.Sprintf("%s.", host)] = ip

	cmd := exec.Command("sudo", "ifconfig", "vlan0", "alias", ip, "netmask", "255.255.255.255")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	endpoints[ip] = addr

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	go handleListener(ln)

	w.Write([]byte(fmt.Sprintf("%s:%s", ip, port)))
}

func handleListener(ln net.Listener) {
	for {
		cn, err := ln.Accept()
		if err != nil {
			fmt.Printf("err = %+v\n", err)
			continue
		}

		go handleConnection(cn)
	}
}

func handleConnection(cn net.Conn) {
	defer cn.Close()

	host, _, err := net.SplitHostPort(cn.LocalAddr().String())
	if err != nil {
		cn.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		return
	}

	ep, ok := endpoints[host]
	if !ok {
		cn.Write([]byte(fmt.Sprintf("no endpoint\n")))
		return
	}

	out, err := net.Dial("tcp", ep)
	if err != nil {
		cn.Write([]byte(fmt.Sprintf("error: %s\n", err)))
	}

	go io.Copy(out, cn)
	io.Copy(cn, out)
}

func deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
}

func listEndpoints(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(endpoints)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(data)
}
