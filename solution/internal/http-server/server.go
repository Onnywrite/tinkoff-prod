package server

import (
	"log/slog"

	"solution/internal/http-server/handler"
	"solution/internal/storage"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	address string
	logger  *slog.Logger
	db      storage.Storage
}

func NewServer(address string, db storage.Storage, logger *slog.Logger) *Server {
	return &Server{
		address: address,
		logger:  logger,
		db:      db,
	}
}

func (s *Server) Start() error {
	e := echo.New()

	e.Use(middleware.Logger(), middleware.Recover())

	e.GET("/api/ping", handler.GetPing)

	s.logger.Info("server has been started", "address", s.address)
	return e.Start(s.address)
}
