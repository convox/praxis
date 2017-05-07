package server

import (
	"github.com/convox/praxis/api"
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

	return nil
}
