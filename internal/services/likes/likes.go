package likes

import (
	"context"
	"log/slog"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
)

type Service struct {
	log *slog.Logger

	saver              LikeSaver
	deleter            LikeDeleter
	likesProvider      LikesProvider
	likesCountProvider LikesCountProvider
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

func New(log *slog.Logger, likeSaver LikeSaver, likeDeleter LikeDeleter,
	likesProvider LikesProvider, likesCountProvider LikesCountProvider) *Service {
	return &Service{
		log:                log,
		saver:              likeSaver,
		deleter:            likeDeleter,
		likesProvider:      likesProvider,
		likesCountProvider: likesCountProvider,
	}
}
