package server

import (
	"net/http"

	d "acm.stupay.com/internal/database"
	"github.com/go-fuego/fuego"
)

func (s *Server) RegisterUserRoutes(api *fuego.Server) {
	fuego.Post(api, "/register", s.RegisterUser)
	fuego.Post(api, "/login", s.LoginUser)
	fuego.Get(api, "/balance", s.GetUserBalance)
}

func (s *Server) RegisterUser(c fuego.ContextWithBody[d.InsertUserArgs]) (d.User, error) {
	body, err := c.Body()
	if err != nil {
		s.logger.Error("internal server error", err)
		return d.User{}, err
	}
	return s.DB.InsertUser(c.Context(), body)
}

func (s *Server) LoginUser(c fuego.ContextWithBody[d.LoginUserArgs]) (bool, error) {
	body, err := c.Body()
	if err != nil {
		http.Error(c.Response(), "Internal Server Error", http.StatusInternalServerError)
		return false, err
	}

	return s.DB.LoginUser(c.Context(), body)
}

func (s *Server) GetUserBalance(c fuego.ContextNoBody) (float32, error) {
	id := c.QueryParamInt("id")
	return s.DB.GetUserBalance(c.Context(), d.GetUserBalanceArgs{UserID: id})
}
