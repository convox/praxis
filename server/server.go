package server

import (
	"github.com/convox/praxis/api"
	"github.com/convox/praxis/server/controllers"
	"github.com/pkg/errors"
)

type Server struct {
	*api.Server
}

func New() *Server {
	server := api.New("api", "rack.convox")

	Routes(server)

	return &Server{Server: server}
}

func (s *Server) Setup() error {
	if err := controllers.Setup(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
