package server

import (
	"log/slog"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/http-server/handler"
	authhandler "github.com/Onnywrite/tinkoff-prod/internal/http-server/handler/auth"
	privatehandler "github.com/Onnywrite/tinkoff-prod/internal/http-server/handler/private"
	mymiddleware "github.com/Onnywrite/tinkoff-prod/internal/http-server/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	address           string
	logger            *slog.Logger
	certPath, keyPath string

	countriesService CountriesService
	usersService     UsersService
	feedService      FeedService
	likesService     LikesService
}

type CountriesService interface {
	handler.CountriesProvider
	handler.CountryProvider
}

type UsersService interface {
	authhandler.UserRegistrator
	authhandler.IdentityProvider
	authhandler.AccessTokenUpdater
	privatehandler.UserProvider
}

type FeedService interface {
	privatehandler.PostCreator
	privatehandler.AllFeedProvider
	privatehandler.AuthorFeedProvider
}

type LikesService interface {
	privatehandler.Liker
	privatehandler.Unliker
	privatehandler.LikesProvider
}

func NewServer(logger *slog.Logger, address, certPath, keyPath string,
	countriesService CountriesService, usersService UsersService, feedService FeedService, likesService LikesService) *Server {
	return &Server{
		logger:           logger,
		address:          address,
		certPath:         certPath,
		keyPath:          keyPath,
		feedService:      feedService,
		countriesService: countriesService,
		likesService:     likesService,
		usersService:     usersService,
	}
}

func (s *Server) Start() error {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{"*"},
	}))

	{
		g := e.Group("api/", mymiddleware.Logger(s.logger), middleware.Recover())

		g.GET("ping", handler.GetPing())
		g.GET("countries", handler.GetCountries(s.countriesService))
		g.GET("countries/:alpha2", handler.GetCountryAlpha(s.countriesService))
		{
			authg := g.Group("auth/")

			authg.POST("register", authhandler.PostRegister(s.usersService))
			authg.POST("sign-in", authhandler.PostSignIn(s.usersService))
			authg.POST("refresh", authhandler.PostRefresh(s.usersService))
		}
		{
			privateg := g.Group("private/", mymiddleware.Authorized())

			privateg.GET("me", privatehandler.GetMe(s.usersService))
			privateg.POST("me/feed", privatehandler.PostMeFeed(s.feedService))
			privateg.GET("feed", privatehandler.GetFeed(s.feedService), mymiddleware.Pagination(100))
			{
				feedg := privateg.Group("posts/", mymiddleware.IdParam("post_id"))

				feedg.GET(":post_id/likes", privatehandler.GetLikes(s.likesService), mymiddleware.Pagination(100))
				feedg.POST(":post_id/like", privatehandler.PostLike(s.likesService))
				feedg.DELETE(":post_id/like", privatehandler.DeleteLike(s.likesService))
			}
			{
				profilesg := privateg.Group("profiles/", mymiddleware.IdParam("user_id"))

				profilesg.GET(":user_id", privatehandler.GetProfile(s.usersService))
				profilesg.GET(":user_id/feed", privatehandler.GetProfileFeed(s.feedService), mymiddleware.Pagination(100))
			}
		}
	}

	s.logger.Info("server has been started", "address", s.address)
	return e.StartTLS(s.address, s.certPath, s.keyPath)
}
