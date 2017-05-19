package api

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/convox/praxis/logger"
)

type Server struct {
	Hostname   string
	Logger     *logger.Logger
	Router     *Router
	middleware []Middleware
}

func (s *Server) Listen(proto, addr string) error {
	s.Logger.At("listen").Logf("hostname=%q addr=%q", s.Hostname, addr)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	switch proto {
	case "http2", "h2", "tcp":
		config := &tls.Config{
			NextProtos: []string{"h2"},
		}

		cert, err := generateSelfSignedCertificate(s.Hostname)
		if err != nil {
			return err
		}

		config.Certificates = append(config.Certificates, cert)

		l = tls.NewListener(l, config)
	}

	return http.Serve(l, s)
}

func (s *Server) Route(method, path string, fn HandlerFunc) {
	s.Router.Route(method, path, fn)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func (s *Server) Subrouter(prefix string) Router {
	return Router{
		Parent: s.Router,
		Router: s.Router.PathPrefix(prefix).Subrouter(),
		Server: s,
	}
}

func (s *Server) Use(mw Middleware) {
	s.Router.Use(mw)
}

func (s *Server) UseHandlerFunc(fn http.HandlerFunc) {
	s.Router.UseHandlerFunc(fn)
}
