package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/types"
)

var (
	Rack *rack.Client
)

func init() {
	r, err := rack.NewFromEnv()
	if err != nil {
		die(err)
	}

	Rack = r
}

func main() {
	if len(os.Args) != 4 {
		usage()
	}

	protocol := os.Args[1]
	style := os.Args[2]
	target := os.Args[3]

	switch style {
	case "redirect":
		if err := handleRedirect(protocol, target); err != nil {
			die(err)
		}
	case "target":
		if err := handleTarget(protocol, target); err != nil {
			die(err)
		}
	default:
		usage()
	}
}

func handleRedirect(protocol, target string) error {
	return http.ListenAndServe(":3000", redirect(target))
}

func redirect(target string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, err := url.Parse(target)
		if err != nil {
			http.Error(w, "could not parse target", 500)
			return
		}

		if t.Scheme == "" {
			t.Scheme = r.URL.Scheme
		}

		if t.Host == "" {
			t.Host = r.Host
		}

		hp := strings.Split(r.Host, ":")

		if t.Hostname() == "" {
			t.Host = hp[0] + t.Host
		}

		if t.Port() == "" && len(hp) > 1 && hp[1] != "" {
			t.Host += fmt.Sprintf(":%s", hp[1])
		}

		if len(r.URL.Path) > 0 {
			t.Path = strings.Replace(t.Path, "*", r.URL.Path[1:], -1)
		}

		http.Redirect(w, r, t.String(), 301)
	})
}

func handleTarget(protocol, target string) error {
	app := os.Getenv("APP")

	u, err := url.Parse(target)
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		return err
	}

	defer ln.Close()

	for {
		cn, err := ln.Accept()
		if err != nil {
			return err
		}

		ps, err := Rack.ProcessList(app, types.ProcessListOptions{Service: u.Hostname()})
		if err != nil {
			return err
		}

		if len(ps) < 1 {
			return fmt.Errorf("no processes for service: %s", u.Hostname())
		}

		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return err
		}

		go Rack.ProxyStart(app, ps[0].Id, port, cn)
	}

	return nil
}

func handleRequest(in net.Conn, service, port string) {
	defer in.Close()

	// out, err := net.Dial("tcp", addr)
	// if err != nil {
	//   fmt.Fprintf(os.Stderr, "error: %s\n", err)
	//   return
	// }

	// go io.Copy(out, in)
	// io.Copy(in, out)
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(1)
}

func usage() {
	die(fmt.Errorf("usage: proxy <protocol> <style> <target>"))
}
