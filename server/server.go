package server

import (
	"github.com/convox/praxis/api"
	"github.com/convox/praxis/server/controllers"
)

type Server struct {
	*api.Server
}

func New() *Server {
	api := api.New("rack.convox")

	api.Route("POST", "/apps", controllers.AppCreate)

	return &Server{Server: api}
}
