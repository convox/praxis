package server

import (
	"github.com/convox/praxis/api"
	"github.com/convox/praxis/server/controllers"
)

type Server struct {
	*api.Server
}

func New() *Server {
	api := api.New("convox", "rack.convox")

	api.Route("POST", "/apps", "apps.create", controllers.AppCreate)

	return &Server{Server: api}
}
