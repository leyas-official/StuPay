package server

import (
	"github.com/go-fuego/fuego"
)

func (s *Server) RegisterHomeRoutes(api *fuego.Server) {
	fuego.Get(api, "/", s.HelloWorldHandler)
}

func (s *Server) HelloWorldHandler(c fuego.ContextNoBody) (string, error) {
	return "Hello World!", nil
}
