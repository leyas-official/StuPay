package server

import (
	"github.com/go-fuego/fuego"
)

func (s *Server) RegisterRoutes() {
	apiRoute := fuego.Group(s.server, "/v1")

	fuego.Use(apiRoute)
	{
		s.RegisterHomeRoutes(apiRoute)
		s.RegisterUserRoutes(apiRoute)
		s.RegisterTransferRoutes(apiRoute)
		s.RegisterBookRoutes(apiRoute)
	}
}
