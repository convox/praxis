package server

import (
	"net/http"
	"os"

	"github.com/convox/api"
	"github.com/convox/praxis/server/controllers"
)

type Server struct {
	*api.Server
}

func New() *Server {
	server := api.New("rack", "convox.rack")

	Routes(server)

	return &Server{Server: server}
}

func (s *Server) Setup() error {
	if err := controllers.Init(); err != nil {
		return err
	}

	if pw := os.Getenv("PASSWORD"); pw != "" {
		s.Use(authenticate(pw))
	}

	return nil
}

func authenticate(password string) api.Middleware {
	return func(fn api.HandlerFunc) api.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, c *api.Context) error {
			key, _, ok := r.BasicAuth()

			if !ok || key != password {
				return api.Errorf(401, "invalid auth")
			}

			return fn(w, r, c)
		}
	}
}
