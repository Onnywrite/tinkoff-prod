package feed

import (
	"context"
	"log/slog"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/services/likes"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
)

type Service struct {
	log *slog.Logger

	d Dependencies
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

type IsLikedProvider interface {
	IsLiked(ctx context.Context, userId, postId uint64) bool
}

type AuthorPostsProvider interface {
	UsersPosts(ctx context.Context, offset, count int, userId uint64) (<-chan models.Post, <-chan ero.Error)
}

type LikesProvider interface {
	Likes(ctx context.Context, opts likes.LikesOptions) (*likes.PagedLikes, ero.Error)
}

type Dependencies struct {
	Provider        PostsProvider
	Counter         PostsCountProvider
	Saver           PostSaver
	AuthorCounter   AuthorPostsCountProvider
	AuthorProvider  AuthorPostsProvider
	IsLikedProvider IsLikedProvider
	LikesProvider   LikesProvider
}

func New(logger *slog.Logger, deps Dependencies) *Service {
	return &Service{
		log: logger,
		d:   deps,
	}
}
