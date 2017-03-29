package cycle

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
)

type HTTP struct {
	Cycles []HTTPCycle
	Server *httptest.Server

	index int
	lock  sync.Mutex
}

type HTTPCycle struct {
	Request  HTTPRequest
	Response HTTPResponse
}

type HTTPRequest struct {
	Method string
	Path   string
	Body   []byte
}

type HTTPResponse struct {
	Code int
	Body []byte
}

func NewHTTP() (*HTTP, error) {
	return &HTTP{Cycles: []HTTPCycle{}}, nil
}

func (s *HTTP) Add(req HTTPRequest, res HTTPResponse) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Cycles = append(s.Cycles, HTTPCycle{Request: req, Response: res})
}

func (s *HTTP) Cycle(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.Cycles) < 1 {
		http.Error(w, "no more cycles", 500)
		return
	}

	cycle := s.Cycles[0]

	s.Cycles = s.Cycles[1:]

	if err := cycle.Request.Match(r); err != nil {
		http.Error(w, err.Error(), 500)
	}

	if cycle.Response.Code > 0 {
		w.WriteHeader(cycle.Response.Code)
	}

	w.Write(cycle.Response.Body)
}

func (s *HTTP) Listen() string {
	s.Server = httptest.NewUnstartedServer(http.HandlerFunc(s.Cycle))

	s.Server.TLS = &tls.Config{
		NextProtos: []string{"h2"},
	}

	s.Server.StartTLS()

	return s.Server.URL
}

func (c *HTTPRequest) Match(r *http.Request) error {
	if err := compare(c.Method, r.Method, "method"); err != nil {
		return err
	}

	if err := compare(c.Path, r.URL.Path, "path"); err != nil {
		return err
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if err := compareb(c.Body, data, "body"); err != nil {
		return err
	}

	return nil
}

func compare(expected, got string, name string) error {
	switch {
	case expected == "":
		return nil
	case expected == got:
		return nil
	default:
		return fmt.Errorf("bad cycle %s: expected:%s got:%s", name, expected, got)
	}
}

func compareb(expected, got []byte, name string) error {
	switch {
	case expected == nil:
		return nil
	case bytes.Compare(expected, got) == 0:
		return nil
	default:
		return fmt.Errorf("bad cycle %s: expected:%q got:%q", name, string(expected), string(got))
	}
}
