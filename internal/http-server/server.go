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

	feedService FeedService
}

type Storage interface {
	handler.CountriesProvider
	handler.CountryProvider
	handler.UserRegistrator
	handler.UserProvider
	handler.UserByIdProvider
}

type FeedService interface {
	handler.PostCreator
	handler.AllFeedProvider
	handler.AuthorFeedProvider
}

func NewServer(address string, db Storage, feedService FeedService, logger *slog.Logger) *Server {
	return &Server{
		address:     address,
		logger:      logger,
		db:          db,
		feedService: feedService,
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
		{
			authg := g.Group("auth/")

			authg.POST("register", handler.PostRegister(s.db))
			authg.POST("sign-in", handler.PostSignIn(s.db))
			authg.POST("refresh", handler.PostRefresh(s.db))
		}
		{
			privateg := g.Group("private/", mymiddleware.Authorized())

			privateg.GET("feed", handler.GetFeed(s.feedService))
			privateg.GET("me", handler.GetMe(s.db))
			privateg.POST("me/feed", handler.PostMeFeed(s.feedService))
			privateg.GET("profiles/:id", handler.GetProfile(s.db))
			privateg.GET("profiles/:id/feed", handler.GetProfileFeed(s.feedService))
		}
	}

	s.logger.Info("server has been started", "address", s.address)
	return e.Start(s.address)
}
