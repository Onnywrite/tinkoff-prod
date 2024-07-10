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
	handler.UserRegistrator
	handler.UserProvider
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

	{
		g := e.Group("api/", mymiddleware.Logger(s.logger), middleware.Recover(), mymiddleware.Cors())

		g.GET("ping", handler.GetPing())
		g.GET("countries", handler.GetCountries(s.db))
		g.GET("countries/:alpha2", handler.GetCountryAlpha(s.db))
		g.POST("auth/register", handler.PostRegister(s.db))
		g.POST("auth/sign-in", handler.PostSignIn(s.db))
		g.GET("me/profile", handler.GetMeProfile(s.db), mymiddleware.Authorized())
	}

	s.logger.Info("server has been started", "address", s.address)
	return e.Start(s.address)
}
