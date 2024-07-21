package likes

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

type LikeSaver interface {
	SaveLike(ctx context.Context, like models.Like) ero.Error
}

type LikeDeleter interface {
	DeleteLike(ctx context.Context, like models.Like) ero.Error
}

type LikesProvider interface {
	Likes(ctx context.Context, offset, count int, postId uint64) (<-chan models.Like, <-chan ero.Error)
}

type LikesCountProvider interface {
	LikesNum(ctx context.Context, postId uint64) (uint64, ero.Error)
}

type LikeProvider interface {
	Like(ctx context.Context, userId, postId uint64) (models.Like, ero.Error)
}

type Dependencies struct {
	Saver        LikeSaver
	Deleter      LikeDeleter
	Provider     LikesProvider
	LikesCounter LikesCountProvider
	LikeProvider LikeProvider
}

func New(log *slog.Logger, deps Dependencies) *Service {
	return &Service{
		log: log,
		d:   deps,
	}
}
