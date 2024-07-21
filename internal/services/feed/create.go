package feed

import (
	"context"
	"errors"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

func (s *Service) CreatePost(ctx context.Context, post NewPost) (uint64, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "feed.Service.CreatePost")

	if err := post.Validate(); err != nil {
		s.log.WarnContext(err.Context(ctx), "post hasn't passed validation")
		return 0, err
	}

	id, err := s.saver.SavePost(ctx, &models.Post{
		Author: models.User{
			Id: post.AuthorId,
		},
		Content:    *post.Content,
		ImagesUrls: models.StringSlice(post.ImagesUrls),
	})
	switch {
	case errors.Is(err, storage.ErrForeignKeyConstraint):
		s.log.WarnContext(err.Context(ctx), "author not found")
		return 0, ero.New(logCtx.WithParent(err.Context(ctx)).Build(), ero.CodeNotFound, ErrAuthorNotFound)
	case err != nil:
		s.log.ErrorContext(err.Context(ctx), "error while saving post")
		return 0, ero.New(logCtx.WithParent(err.Context(ctx)).With("error", err).Build(), ero.CodeInternal, ErrInternal)
	}

	return id, err
}
