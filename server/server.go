package server

import (
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

func Setup() error {
	if err := controllers.Init(); err != nil {
		return err
	}

	return nil
}
