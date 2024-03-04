package server

import (
	"log/slog"

	"solution/internal/http-server/handler"
	mymiddleware "solution/internal/http-server/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	address string
	logger  *slog.Logger
	db      Storage
}

type Storage interface {
	handler.CountriesProvider
	handler.CountryProvider
}

func NewServer(address string, db Storage, logger *slog.Logger) *Server {
	return &Server{
		address: address,
		logger:  logger,
		db:      db,
	}
}

func (s *Server) Start() error {
	e := echo.New()

	e.Use(mymiddleware.Logger(s.logger), middleware.Recover())

	e.GET("/api/ping", handler.GetPing())
	e.GET("/api/countries", handler.GetCountries(s.db))
	e.GET("/api/countries/:alpha2", handler.GetCountryAlpha(s.db))

	s.logger.Info("server has been started", "address", s.address)
	return e.Start(s.address)
}
