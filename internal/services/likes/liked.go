package likes

import (
	"context"
)

func (s *Service) IsLiked(ctx context.Context, userId, postId uint64) bool {
	// logCtx := erolog.NewContextBuilder().With("op", "likes.Service.IsLiked").With("user_id", userId).With("post_id", postId)

	_, err := s.d.LikeProvider.Like(ctx, userId, postId)
	if err != nil {
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
