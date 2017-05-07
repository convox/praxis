package api

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/websocket"

	"github.com/convox/praxis/types"
	"github.com/gorilla/mux"
)

type Router struct {
	*mux.Router
	Middleware []Middleware
	Parent     *Router
	Server     *Server
}

func (rt *Router) Route(name, method, path string, fn HandlerFunc) {
	rt.Handle(path, rt.api(name, fn)).Methods(method)
}

func (rt *Router) Stream(name, path string, fn StreamFunc) {
	rt.Handle(path, rt.streamWebsocket(name, fn)).Methods("GET").Headers("Connection", "Upgrade")
	rt.Handle(path, rt.streamHTTP2(name, fn)).Methods("POST")
}

func (rt *Router) Use(mw Middleware) {
	rt.Middleware = append(rt.Middleware, mw)
}

func (rt *Router) UseHandlerFunc(fn http.HandlerFunc) {
	rt.Middleware = append(rt.Middleware, func(gn HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			fn(w, r)
			return gn(w, r, c)
		}
	})
}

func (rt *Router) streamHTTP2(at string, fn StreamFunc) http.HandlerFunc {
	return rt.api(at, func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return fn(types.Stream{Reader: r.Body, Writer: w}, c)
	})
}

func (rt *Router) streamWebsocket(at string, fn StreamFunc) websocket.Handler {
	return func(ws *websocket.Conn) {
		c, err := rt.context(at, ws, ws.Request())
		if err != nil {
			fmt.Printf("err = %+v\n", err)
			return
		}

		if err := fn(ws, c); err != nil {
			fmt.Printf("err = %+v\n", err)
			return
		}
	}
}

func (rt *Router) api(at string, fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := rt.context(at, w, r)
		if err != nil {
			e := fmt.Errorf("context error: %s", err)
			c.LogError(e)
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}

		c.Start("method=%q path=%q", r.Method, r.URL.Path)

		mw := []Middleware{}

		if rt.Parent != nil {
			mw = append(mw, rt.Parent.Middleware...)
		}

		mw = append(mw, rt.Middleware...)

		fnmw := rt.wrap(fn, mw...)

		switch err := fnmw(w, r, c).(type) {
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

func (rt *Router) context(name string, w io.Writer, r *http.Request) (*Context, error) {
	id, err := Key(12)
	if err != nil {
		return nil, err
	}

	return &Context{
		logger:  rt.Server.Logger.Namespace("id=%s route=%s", id, name),
		request: r,
		writer:  w,
	}, nil
}

func (rt *Router) wrap(fn HandlerFunc, m ...Middleware) HandlerFunc {
	if len(m) == 0 {
		return fn
	}

	return m[0](rt.wrap(fn, m[1:len(m)]...))
}
