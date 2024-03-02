package server

import (
	"log/slog"

	"solution/internal/http-server/handler"

	"github.com/labstack/echo/v4"
)

type Server struct {
	address string
	logger  *slog.Logger
}

func NewServer(address string, logger *slog.Logger) *Server {
	return &Server{
		address: address,
		logger:  logger,
	}
}

func (s *Server) Start() error {
	e := echo.New()

	e.GET("/api/ping", handler.GetPing)

	s.logger.Info("server has been started", "address", s.address)

	e.Start(s.address)

	return nil
}
