package likes

import (
	"context"
	"errors"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

func (s *Service) Like(ctx context.Context, userId, postId uint64) ero.Error {
	logCtx := erolog.NewContextBuilder().With("op", "likes.Service.Like").With("user_id", userId).With("post_id", postId)

	err := s.d.Saver.SaveLike(ctx, models.Like{
		User: models.User{
			Id: userId,
		},
		Post: models.Post{
			Id: postId,
		},
	})
	switch {
	case errors.Is(err, storage.ErrForeignKeyConstraint):
		s.log.DebugContext(logCtx.BuildContext(), "user or post does not exist")
		return ero.New(logCtx.WithParent(err.Context(ctx)).Build(), ero.CodeNotFound, ErrNotFound)
	case errors.Is(err, storage.ErrUniqueConstraint):
		s.log.DebugContext(logCtx.BuildContext(), "already liked")
		return ero.New(logCtx.WithParent(err.Context(ctx)).Build(), ero.CodeExists, ErrAlreadyLiked)
	case err != nil:
		s.log.ErrorContext(err.Context(ctx), "error while saving like")
		return ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, ErrInternal)
	}

	return nil
}

func (s *Service) Unlike(ctx context.Context, userId, postId uint64) ero.Error {
	logCtx := erolog.NewContextBuilder().With("op", "likes.Service.Like").With("user_id", userId).With("post_id", postId)

	err := s.d.Deleter.DeleteLike(ctx, models.Like{
		User: models.User{
			Id: userId,
		},
		Post: models.Post{
			Id: postId,
		},
	})
	switch {
	case errors.Is(err, storage.ErrNoRows):
		s.log.DebugContext(logCtx.BuildContext(), "post has not been liked")
		return ero.New(logCtx.WithParent(err.Context(ctx)).Build(), ero.CodeNotFound, ErrAlreadyUnliked)
	case err != nil:
		s.log.ErrorContext(err.Context(ctx), "error while saving like")
		return ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, ErrInternal)
	}

	return nil
}
