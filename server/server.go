package server

import (
	"fmt"
	"net/http"

	"github.com/convox/api"
)

type Server struct {
	*api.Server
}

func New() *Server {
	server := api.New("rack", "convox.rack")

	server.UseHandlerFunc(mw1)

	Routes(server)

	return &Server{Server: server}
}

func mw1(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("bar"))
	fmt.Println("mw1")
}

func mw2(fn api.HandlerFunc) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, c *api.Context) error {
		fmt.Println("mw2 before")
		// return fmt.Errorf("whaaa")
		if err := fn(w, r, c); err != nil {
			return err
		}
		fmt.Println("mw2 after")
		return nil
	}
}
