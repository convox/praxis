package server

import (
	"github.com/convox/praxis/api"
	"github.com/convox/praxis/server/controllers"
)

type Server struct {
	*api.Server
}

func New() *Server {
	api := api.New("convox.rack", "convox.rack")

	api.Route("app.create", "POST", "/apps", controllers.AppCreate)

	api.Route("object.store", "POST", "/apps/{app}/objects/{key:.*}", controllers.ObjectStore)

	return &Server{Server: api}
}
