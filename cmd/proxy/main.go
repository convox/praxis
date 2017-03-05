package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

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

	switch protocol {
	case "https", "tls":
		cert, err := generateSelfSignedCertificate("convox.local")

		if err != nil {
			return err
		}

		ln = tls.NewListener(ln, &tls.Config{
			Certificates: []tls.Certificate{cert},
		})
	}

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

		switch u.Scheme {
		case "https", "tls":
			r, w := net.Pipe()

			tc := tls.Client(w, &tls.Config{
				InsecureSkipVerify: true,
			})

			go io.Copy(cn, tc)
			go io.Copy(tc, cn)

			cn = r
		}

		go Rack.ProxyStart(app, ps[0].Id, port, cn)
	}
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(1)
}

func usage() {
	die(fmt.Errorf("usage: proxy <protocol> <style> <target>"))
}

func generateSelfSignedCertificate(host string) (tls.Certificate, error) {
	rkey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return tls.Certificate{}, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"convox"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{host},
	}

	data, err := x509.CreateCertificate(rand.Reader, &template, &template, &rkey.PublicKey, rkey)

	if err != nil {
		return tls.Certificate{}, err
	}

	pub := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: data})
	key := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rkey)})

	return tls.X509KeyPair(pub, key)
}
