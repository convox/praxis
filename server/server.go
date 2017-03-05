package server

import "github.com/convox/api"

type Server struct {
	*api.Server
}

func New() *Server {
	api := api.New("rack", "convox.rack")

	Routes(api)

	return &Server{Server: api}
}
