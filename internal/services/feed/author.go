package feed

import (
	"context"
	"time"

	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

type AuthorFeedOptions struct {
	Page       uint64
	PageSize   uint64
	UserId     uint64
	LikesCount uint64
	FormatDate func(time.Time) string
}

func (s *Service) AuthorFeed(ctx context.Context, opts AuthorFeedOptions) (*PagedProfileFeed, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "feed.Service.AllFeed").With("opts.Page", opts.Page).With("page_size", opts.PageSize).With("user_id", opts.UserId)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	postsCh, errCh := s.d.AuthorProvider.UsersPosts(ctx, int(opts.Page-1)*int(opts.PageSize), int(opts.PageSize), opts.UserId)

	postsCount, eroErr := s.d.AuthorCounter.UsersPostsNum(ctx, opts.UserId)
	if eroErr != nil {
		s.log.ErrorContext(eroErr.Context(ctx), "internal error")
		return nil, eroErr
	}

	posts := make([]LikedAuthorlessPost, 0, opts.PageSize)
	for p := range postsCh {
		likesInfo, eroErr := s.getLikesForPost(ctx, p.Id, opts.UserId, opts.LikesCount, opts.FormatDate)
		if eroErr != nil {
			s.log.ErrorContext(eroErr.Context(ctx), "error while getting likes for post in profile")
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

		posts = append(posts, LikedAuthorlessPost{
			Liked:      likesInfo.isLiked,
			LikesCount: likesInfo.count,
			Likes:      likesInfo.likes,
			AuthorlessPost: AuthorlessPost{
				Id:          p.Id,
				Content:     p.Content,
				ImageUrl:    url,
				PublishedAt: opts.FormatDate(p.PublishedAt),
				UpdatedAt:   updatedAt,
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

	return &PagedProfileFeed{
		First:   1,
		Current: uint64(opts.Page),
		Last:    (postsCount + opts.PageSize - 1) / opts.PageSize,
		Posts:   posts,
	}, nil
}
