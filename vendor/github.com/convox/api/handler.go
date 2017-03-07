package api

import "net/http"

type HandlerFunc func(w http.ResponseWriter, r *http.Request, c *Context) error
type Middleware func(fn HandlerFunc) HandlerFunc

func NewHandlerFunc(fn http.HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		fn(w, r)
		return nil
	}
}
