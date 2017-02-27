package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func main() {
	if len(os.Args) != 4 {
		usage()
	}

	protocol := os.Args[1]
	style := os.Args[2]
	target := os.Args[3]

	fmt.Printf("protocol = %+v\n", protocol)
	fmt.Printf("style = %+v\n", style)
	fmt.Printf("target = %+v\n", target)

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

	fmt.Printf("os.Args = %+v\n", os.Args)
}

func handleRedirect(protocol, target string) error {
	handler, err := redirect(target)
	if err != nil {
		return err
	}

	return http.ListenAndServe(":3000", handler)
}

func redirect(target string) (http.Handler, error) {
	t, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rdr := ""

		fmt.Printf("t = %#v\n", t)

		if t.Scheme != "" {
			rdr += fmt.Sprintf("%s://", t.Scheme)
		}

		if h := t.Hostname(); h != "" {
			rdr += h
		}

		if p := t.Port(); p != "" {
			rdr += fmt.Sprintf(":%s", p)
		}

		if p := t.Path; p != "" {
			rdr += p
		} else {
			rdr += r.URL.Path
		}

		fmt.Printf("rdr = %+v\n", rdr)
		// proto := r
		// rdr := fmt.Sprintf("%s://%s:%s%s", proto, host, port, path)
		// fmt.Printf("rdr = %+v\n", rdr)
		// host := r.Host
		// fmt.Printf("w = %+v\n", w)
		// fmt.Printf("r = %+v\n", r)
	}), nil
}

func handleTarget(protocol, target string) error {
	fmt.Printf("protocol = %+v\n", protocol)
	fmt.Printf("target = %+v\n", target)
	return nil
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(1)
}

func usage() {
	die(fmt.Errorf("usage: proxy <protocol> <style> <target>"))
}
