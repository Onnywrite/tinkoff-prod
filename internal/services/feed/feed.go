package feed

import (
	"context"
	"log/slog"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
)

type Service struct {
	log *slog.Logger

	provider       PostsProvider
	countProvider  PostsCountProvider
	saver          PostSaver
	aCountProvider AuthorPostsCountProvider
	aProvider      AuthorPostsProvider
}

type PostsProvider interface {
	Posts(ctx context.Context, offset, count int) (<-chan models.Post, <-chan ero.Error)
}

type PostsCountProvider interface {
	PostsNum(context.Context) (uint64, ero.Error)
}

type PostSaver interface {
	SavePost(ctx context.Context, post *models.Post) (uint64, ero.Error)
}

type AuthorPostsCountProvider interface {
	UsersPostsNum(ctx context.Context, userId uint64) (uint64, ero.Error)
}

type AuthorPostsProvider interface {
	UsersPosts(ctx context.Context, offset, count int, userId uint64) (<-chan models.Post, <-chan ero.Error)
}

func New(logger *slog.Logger, provider PostsProvider, countProvider PostsCountProvider,
	saver PostSaver, authorProvider AuthorPostsProvider, authorCountProvider AuthorPostsCountProvider) *Service {
	return &Service{
		log:            logger,
		provider:       provider,
		countProvider:  countProvider,
		saver:          saver,
		aCountProvider: authorCountProvider,
		aProvider:      authorProvider,
	}
}
