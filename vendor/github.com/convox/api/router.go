package api

import (
	"fmt"
	"net/http"

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

func (rt *Router) context(name string, w http.ResponseWriter, r *http.Request) (*Context, error) {
	id, err := Key(12)
	if err != nil {
		return nil, err
	}

	return &Context{
		logger:   rt.Server.Logger.Namespace("id=%s route=%s", id, name),
		request:  r,
		response: w,
	}, nil
}

func (rt *Router) wrap(fn HandlerFunc, m ...Middleware) HandlerFunc {
	if len(m) == 0 {
		return fn
	}

	return m[0](rt.wrap(fn, m[1:len(m)]...))
}
