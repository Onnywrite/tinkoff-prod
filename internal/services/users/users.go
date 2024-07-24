package users

import (
	"context"
	"log/slog"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
)

type Service struct {
	log *slog.Logger

	d Dependencies
}

type UserByIdProvider interface {
	UserById(ctx context.Context, id uint64) (*models.User, ero.Error)
}

type UserByEmailProvider interface {
	UserByEmail(ctx context.Context, email string) (*models.User, ero.Error)
}

type UserSaver interface {
	SaveUser(ctx context.Context, user *models.User) (*models.User, ero.Error)
}

type Dependencies struct {
	ByIdProvider    UserByIdProvider
	ByEmailProvider UserByEmailProvider
	Saver           UserSaver
}

func New(log *slog.Logger, deps Dependencies) *Service {
	return &Service{
		log: log,
		d:   deps,
	}
}
