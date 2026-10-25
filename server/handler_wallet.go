package server

import (
	ws "acm.stupay.com/internal/wallet"
	"github.com/go-fuego/fuego"
)

func (s *Server) RegisterTransferRoutes(api *fuego.Server) {
	fuego.Get(api, "/transfer", s.TransferMoneyHandler)
}

func (s *Server) TransferMoneyHandler(c fuego.ContextWithBody[ws.TransferMoneyArgs]) (bool, error) {
	ctx := c.Context()
	body, err := c.Body()
	if err != nil {
		s.logger.Error("internal server error", err)
		return false, err
	}

	return s.ws.TransferMoney(ctx, body)
}
