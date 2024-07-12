package server

import (
	"log/slog"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/http-server/handler"
	mymiddleware "github.com/Onnywrite/tinkoff-prod/internal/http-server/middleware"

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
	handler.UserByIdProvider
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

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	{
		g := e.Group("api/", mymiddleware.Logger(s.logger), middleware.Recover())

		g.GET("ping", handler.GetPing())
		g.GET("countries", handler.GetCountries(s.db))
		g.GET("countries/:alpha2", handler.GetCountryAlpha(s.db))
		g.GET("me", handler.GetMeProfile(s.db), mymiddleware.Authorized())
		{
			authg := g.Group("auth/")

			authg.POST("register", handler.PostRegister(s.db))
			authg.POST("sign-in", handler.PostSignIn(s.db))
			authg.POST("refresh", handler.PostRefresh(s.db))
		}
	}

	s.logger.Info("server has been started", "address", s.address)
	return e.Start(s.address)
}
