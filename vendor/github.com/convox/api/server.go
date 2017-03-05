package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/convox/logger"
	"github.com/gorilla/mux"
)

type Server struct {
	Hostname string
	Router   *mux.Router
	logger   *logger.Logger
}

func (s *Server) Listen(proto, addr string) error {
	s.logger.At("listen").Logf("hostname=%q addr=%q", s.Hostname, addr)

	l, err := net.Listen(proto, addr)
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

	return http.Serve(tls.NewListener(l, config), s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func (s *Server) Route(name, method, path string, fn HandlerFunc) {
	s.Router.Handle(path, s.api(name, fn)).Methods(method)
}

func (s *Server) api(at string, fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := s.context(at, w, r)
		if err != nil {
			e := fmt.Errorf("context error: %s", err)
			c.LogError(e)
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}

		c.Start("method=%q path=%q", r.Method, r.URL.Path)

		switch err := fn(w, r, c).(type) {
		case Error:
			c.LogError(err)
			http.Error(w, err.Error(), err.Code)
		case error:
			c.LogError(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		case nil:
			c.LogSuccess()
		default:
			err = fmt.Errorf("invalid controller return")
			c.LogError(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) context(name string, w http.ResponseWriter, r *http.Request) (*Context, error) {
	id, err := Key(12)
	if err != nil {
		return nil, err
	}

	return &Context{
		logger:   s.logger.Namespace("id=%s route=%s", id, name),
		request:  r,
		response: w,
	}, nil
}
