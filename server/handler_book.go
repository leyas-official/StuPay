package server

import (
	ws "acm.stupay.com/internal/wallet"
	"github.com/go-fuego/fuego"
)

func (s *Server) RegisterBookRoutes(api *fuego.Server) {
	fuego.Post(api, "/reserve", s.ReserveBookHandler)
}

func (s *Server) ReserveBookHandler(c fuego.ContextWithBody[ws.BorrowBookArgs]) (bool, error) {
	ctx := c.Context()
	body, err := c.Body()
	if err != nil {
		s.logger.Error("internal server error", err)
		return false, err
	}

	print(body.LibraryID)

	return s.ws.BorrowBook(ctx, body)
}
