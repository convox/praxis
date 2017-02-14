package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"

	"github.com/convox/logger"
	"github.com/gorilla/mux"
)

type Context struct {
	logger  *logger.Logger
	request *http.Request
}

type Error struct {
	error
	Code int
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request, c *Context) error

type Server struct {
	Hostname string
	Router   *mux.Router
	logger   *logger.Logger
}

func New(ns, hostname string) *Server {
	return &Server{
		Hostname: hostname,
		Router:   mux.NewRouter(),
		logger:   logger.New(fmt.Sprintf("ns=%s", ns)),
	}
}

func Errorf(code int, format string, args ...interface{}) Error {
	return Error{
		error: fmt.Errorf(format, args...),
		Code:  code,
	}
}

func (c *Context) LogError(err error) {
	_, file, line, _ := runtime.Caller(1)
	location := fmt.Sprintf("%s:%d", file, line)

	log := c.logger.At("end")

	switch t := err.(type) {
	case Error:
		switch t.Code / 100 {
		case 4:
			log.Logf("state=error type=user code=%d error=%q location=%q", t.Code, t.Error(), location)
		case 5:
			log.Logf("state=error type=server code=%d error=%q location=%q", t.Code, t.Error(), location)
		default:
			log.Logf("state=error type=unknown code=%d error=%q location=%q", t.Code, t.Error(), location)
		}
	case error:
		log.Logf("state=error code=500 error=%q location=%q", t.Error(), location)
	case nil:
	default:
		log.Logf("state=error code=500 error=%q location=%q", "unknown error type", location)
	}
}

func (c *Context) LogParams(names ...string) {
	params := make([]string, len(names))

	for i, name := range names {
		params[i] = fmt.Sprintf("%s=%q", name, c.request.FormValue(name))
	}

	c.logger.At("params").Logf(strings.Join(params, " "))
}

func (c *Context) LogSuccess() {
	c.logger.At("end").Success()
}

func (c *Context) Logf(format string, args ...interface{}) {
	c.logger.Logf(format, args...)
}

func (c *Context) Start(format string, args ...interface{}) {
	c.logger = c.logger.Start()
	c.logger.At("start").Logf(format, args...)
}

func (c *Context) Tag(format string, args ...interface{}) {
	c.logger = c.logger.Namespace(format, args...)
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

func (s *Server) Route(method, path, name string, fn HandlerFunc) {
	s.Router.Handle(path, s.api(name, fn)).Methods(method)
}

func (s *Server) api(at string, fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := s.context(at, w, r)

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

func (s *Server) context(name string, w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		logger:  s.logger.Namespace("route=%s id=1234", name),
		request: r,
	}
}
