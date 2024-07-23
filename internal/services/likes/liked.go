package likes

import (
	"context"
	"errors"

	"github.com/Onnywrite/tinkoff-prod/internal/storage"
)

func (s *Service) IsLiked(ctx context.Context, userId, postId uint64) bool {
	_, err := s.d.LikeProvider.Like(ctx, userId, postId)
	switch {
	case errors.Is(err, storage.ErrNoRows):
		s.log.DebugContext(err.Context(ctx), "no likes")
		return false
	case err != nil:
		s.log.ErrorContext(err.Context(ctx), "error while getting like")
		return false
	}

	return true
}

type IDistributedCache interface {
}

type FuncCtx[TQuery any, TResponse any] func(context.Context, TQuery) (TResponse, error)

func GetOrCreateAsync[T any, TQuery any, TResponse any](
	ctx context.Context,
	cache IDistributedCache,
	key string,
	fact FuncCtx[TQuery, TResponse],
) error {
	return nil
}
