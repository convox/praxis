package server

import "github.com/convox/api"

type Server struct {
	*api.Server
}

func New() *Server {
	server := api.New("rack", "convox.rack")

	Routes(server)

	return &Server{Server: server}
}
