package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/Onnywrite/tinkoff-prod/internal/config"
	server "github.com/Onnywrite/tinkoff-prod/internal/http-server"
	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/storage/pg"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

type Application struct {
	log *slog.Logger
	cfg *config.Config
	db  *pg.PgStorage
	srv *server.Server
}

func New(cfg *config.Config) *Application {
	logger := slog.New(erolog.New(os.Stdout, cfg.MustErologConfig()))

	return &Application{
		log: logger,
		cfg: cfg,
	}
}

func (a *Application) MustStart(ctx context.Context) {
	if err := a.Start(ctx); err != nil {
		a.log.Error("could not start", "error", err)
		panic(err)
	}
}

func (a *Application) Start(ctx context.Context) (err error) {
	a.cfg.StartWatch(ctx, a.updateConfig)
	a.db, err = pg.New(a.cfg.Conn)
	if err != nil {
		return err
	}

	relativePath := a.cfg.Dir() + "/"

	a.srv = server.NewServer(a.log, a.db, fmt.Sprintf(":%d", a.cfg.Https.Port), relativePath+a.cfg.Https.Cert, relativePath+a.cfg.Https.Key)
	a.srv.Start()

	a.log.Info("started")
	return nil
}

func (a *Application) MustStop() {
	if err := a.Stop(); err != nil {
		a.log.Error("could not stop", "error", err)
		panic(err)
	}
}

func (a *Application) Stop() error {
	a.log.Info("stopping")
	if a.db == nil {
		return fmt.Errorf("app.Application: database is not initialized")
	}
	if err := a.db.Disconnect(); err != nil {
		a.log.Error("could not disconnect from database", "error", err)
		return err
	}
	a.log.Info("finished")
	return nil
}

func (a *Application) updateConfig(cfg config.Config) {
	if ero.CurrentService != cfg.ServiceName {
		a.cfg.ServiceName = cfg.ServiceName
		ero.CurrentService = cfg.ServiceName
		a.log.Debug("updated service name")
	}

	if a.cfg.AccessToken.Secret != cfg.AccessToken.Secret {
		a.cfg.AccessToken.Secret = cfg.AccessToken.Secret
		secret, err := getSecret(a.cfg.Dir(), cfg.AccessToken.Secret)
		if err != nil {
			a.log.Error("could not update access secret", slog.String("error", err.Error()))
		} else {
			tokens.AccessSecret = secret
			a.log.Debug("updated access secret")
		}
	}
	if a.cfg.RefreshToken.Secret != cfg.RefreshToken.Secret {
		a.cfg.RefreshToken.Secret = cfg.RefreshToken.Secret
		secret, err := getSecret(a.cfg.Dir(), cfg.RefreshToken.Secret)
		if err != nil {
			a.log.Error("could not update refresh secret", slog.String("error", err.Error()))
		} else {
			tokens.RefreshSecret = secret
			a.log.Debug("updated refresh secret")
		}
	}

	if tokens.AccessTTL != cfg.AccessToken.TTL {
		a.cfg.AccessToken.TTL = cfg.AccessToken.TTL
		tokens.AccessTTL = cfg.AccessToken.TTL
		a.log.Debug("updated access ttl")
	}
	if tokens.RefreshTTL != cfg.RefreshToken.TTL {
		a.cfg.RefreshToken.TTL = cfg.RefreshToken.TTL
		tokens.RefreshTTL = cfg.RefreshToken.TTL
		a.log.Debug("updated refresh ttl")
	}

	a.cfg.ResetWatchFreq(cfg.WatchFreq)

	if erologger, ok := a.log.Handler().(*erolog.Logger); ok {
		erologger.UpdateConfig(cfg.MustErologConfig())
	}
	a.log.Debug("updated config")
}

func getSecret(relativePath, secretSomething string) ([]byte, error) {
	secret, isFile := strings.CutPrefix(secretSomething, "file://")
	if isFile {
		secretBytes, err := os.ReadFile(relativePath + "/" + strings.TrimPrefix(secret, "./"))
		if err != nil {
			return nil, err
		}
		return secretBytes, nil
	}

	return []byte(secret), nil
}
