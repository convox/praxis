package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

type Server struct {
	Hostname string
	Router   *mux.Router
}

func New(hostname string) *Server {
	return &Server{
		Hostname: hostname,
		Router:   mux.NewRouter(),
	}
}

func (s *Server) Listen(addr, port string) error {
	l, err := net.Listen(addr, port)
	if err != nil {
		return err
	}

	config := &tls.Config{
		NextProtos: []string{"h2"},
	}

	cert, err := generateSelfSignedCertificate(s.Hostname)
	if err != nil {
		return err
	}

	config.Certificates = append(config.Certificates, cert)

	return http.Serve(tls.NewListener(l, config), s.Router)
}

func (s *Server) Route(method, path string, fn HandlerFunc) {
	s.Router.Handle(path, api(fn)).Methods(method)
}

func api(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		fmt.Printf("err = %+v\n", err)
	}
}
