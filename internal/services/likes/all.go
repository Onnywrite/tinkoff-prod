package likes

import (
	"context"
	"time"

	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

type LikesOptions struct {
	Page       uint64
	PageSize   uint64
	PostId     uint64
	FormatDate func(time.Time) string
}

func (s *Service) Likes(ctx context.Context, opts LikesOptions) (*PagedLikes, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "likes.Service.GetLiked").With("post_id", opts.PostId).With("page", opts.Page).With("page_size", opts.PageSize)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	likesCh, eroCh := s.d.Provider.Likes(ctx, int((opts.Page-1)*opts.PageSize), int(opts.PageSize), opts.PostId)

	likesCount, eroErr := s.d.LikesCounter.LikesNum(ctx, opts.PostId)
	if eroErr != nil {
		s.log.ErrorContext(eroErr.Context(ctx), "error while getting likes count")
		return nil, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeInternal, ErrInternal)
	}

	likes := make([]Like, 0, opts.PageSize)
	for like := range likesCh {
		likes = append(likes, Like{
			User: User{
				Id:       like.User.Id,
				Name:     like.User.Name,
				Lastname: like.User.Lastname,
				Image:    like.User.Image,
			},
			LikedAt: opts.FormatDate(like.LikedAt),
		})
	}

	if err := <-eroCh; err != nil {
		s.log.ErrorContext(err.Context(ctx), "error while getting likers")
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, ErrInternal)
	}

	if len(likes) == 0 {
		s.log.WarnContext(logCtx.BuildContext(), "no likes")
		return nil, ero.New(logCtx.Build(), ero.CodeNotFound, ErrNoLikes)
	}

	return &PagedLikes{
		First:   1,
		Current: opts.Page,
		Last:    (likesCount + opts.PageSize - 1) / opts.PageSize,
		Count:   likesCount,
		Likes:   likes,
	}, nil
}
