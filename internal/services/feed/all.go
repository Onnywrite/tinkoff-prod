package feed

import (
	"context"
	"time"

	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

func (s *Service) AllFeed(ctx context.Context, page, pageSize uint64, formatDate func(time.Time) string) (*PagedFeed, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "feed.Service.AllFeed").With("page", page).With("page_size", pageSize)
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()
	postsCh, errCh := s.provider.Posts(ctx, int(page-1)*int(pageSize), int(pageSize))

	postsCount, eroErr := s.countProvider.PostsNum(ctx)
	if eroErr != nil {
		s.log.ErrorContext(eroErr.Context(ctx), "internal error")
		return nil, eroErr
	}

	posts := make([]FullPost, 0, pageSize)
	for p := range postsCh {
		var url *string
		if p.ImagesUrls == nil || len(p.ImagesUrls) == 0 {
			url = nil
		} else {
			url = &p.ImagesUrls[0]
		}

		var updatedAt *string
		if p.UpdatedAt != nil {
			formatted := formatDate(*p.UpdatedAt)
			updatedAt = &formatted
		} else {
			updatedAt = nil
		}

		posts = append(posts, FullPost{
			Id:          p.Id,
			Content:     p.Content,
			ImageUrl:    url,
			PublishedAt: formatDate(p.PublishedAt),
			UpdatedAt:   updatedAt,
			Author: Author{
				Id:       p.Author.Id,
				Name:     p.Author.Name,
				Lastname: p.Author.Lastname,
				Image:    p.Author.Image,
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
		Current: uint64(page),
		Last:    (postsCount + pageSize - 1) / pageSize,
		Posts:   posts,
	}, nil
}
