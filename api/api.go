package api

import (
	"fmt"
	"net/http"

	"github.com/convox/logger"
	"github.com/convox/praxis/provider/types"
	"github.com/gorilla/mux"
)

type Error struct {
	error
	Code int
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request, c *Context) error

func New(ns, hostname string) *Server {
	logger := logger.New(fmt.Sprintf("ns=%s", ns))
	router := mux.NewRouter()

	server := &Server{
		Hostname: hostname,
		Router:   router,
		logger:   logger,
	}

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := types.Key(12)
		logger.Logf("route=unknown id=%s code=404 method=%q path=%q", id, r.Method, r.URL.Path)
	})

	return server
}

func Errorf(code int, format string, args ...interface{}) Error {
	return Error{
		error: fmt.Errorf(format, args...),
		Code:  code,
	}
}
