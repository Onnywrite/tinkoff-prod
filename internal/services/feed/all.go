package feed

import (
	"context"
	"errors"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/services/likes"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

type AllFeedOptions struct {
	Page       uint64
	PageSize   uint64
	UserId     uint64
	LikesCount uint64
	FormatDate func(time.Time) string
}

type postLikes struct {
	likes   []likes.Like
	count   uint64
	isLiked bool
}

func (s *Service) getLikesForPost(ctx context.Context, postId, userId, maxCount uint64, formatDate func(time.Time) string) (info postLikes, err ero.Error) {
	likesPage, eroErr := s.d.LikesProvider.Likes(ctx, likes.LikesOptions{
		Page:       1,
		PageSize:   maxCount,
		PostId:     postId,
		FormatDate: formatDate,
	})
	switch {
	case errors.Is(eroErr, likes.ErrNoLikes):
		return postLikes{
			likes:   []likes.Like{},
			count:   0,
			isLiked: false,
		}, nil
	case eroErr != nil:
		return postLikes{}, eroErr
	}

	isLiked := s.d.IsLikedProvider.IsLiked(ctx, userId, postId)
	return postLikes{
		likes:   likesPage.Likes,
		count:   likesPage.Count,
		isLiked: isLiked,
	}, nil
}

// refactor: use dynamic schema (map[string]any) and decorator pattern
func (s *Service) AllFeed(ctx context.Context, opts AllFeedOptions) (*PagedFeed, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "feed.Service.AllFeed").With("page", opts.Page).With("page_size", opts.PageSize)
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()
	postsCh, errCh := s.d.Provider.Posts(ctx, int(opts.Page-1)*int(opts.PageSize), int(opts.PageSize))

	postsCount, eroErr := s.d.Counter.PostsNum(ctx)
	if eroErr != nil {
		s.log.ErrorContext(eroErr.Context(ctx), "internal error")
		return nil, eroErr
	}

	posts := make([]LikedPost, 0, opts.PageSize)
	for p := range postsCh {
		likesInfo, eroErr := s.getLikesForPost(ctx, p.Id, opts.UserId, opts.LikesCount, opts.FormatDate)
		if eroErr != nil {
			s.log.ErrorContext(eroErr.Context(ctx), "error while getting likes for post in feed")
			continue
		}

		var url *string
		if p.ImagesUrls == nil || len(p.ImagesUrls) == 0 {
			url = nil
		} else {
			url = &p.ImagesUrls[0]
		}

		var updatedAt *string
		if p.UpdatedAt != nil {
			formatted := opts.FormatDate(*p.UpdatedAt)
			updatedAt = &formatted
		} else {
			updatedAt = nil
		}

		posts = append(posts, LikedPost{
			Liked:      likesInfo.isLiked,
			LikesCount: likesInfo.count,
			Likes:      likesInfo.likes,
			Post: Post{
				Id:          p.Id,
				Content:     p.Content,
				ImageUrl:    url,
				PublishedAt: opts.FormatDate(p.PublishedAt),
				UpdatedAt:   updatedAt,
				Author: Author{
					Id:       p.Author.Id,
					Name:     p.Author.Name,
					Lastname: p.Author.Lastname,
					Image:    p.Author.Image,
				},
			},
		})
	}

	if eroErr = <-errCh; eroErr != nil {
		s.log.ErrorContext(eroErr.Context(ctx), "error while getting posts")
		return nil, ero.New(logCtx.WithParent(eroErr.Context(ctx)).With("error", eroErr).Build(), ero.CodeInternal, ErrInternal)
	}

	if len(posts) == 0 {
		return nil, ero.New(logCtx.Build(), ero.CodeNotFound, ErrNoPosts)
	}

	return &PagedFeed{
		First:   1,
		Current: uint64(opts.Page),
		Last:    (postsCount + opts.PageSize - 1) / opts.PageSize,
		Posts:   posts,
	}, nil
}
