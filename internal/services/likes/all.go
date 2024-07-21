package likes

import (
	"context"
	"time"

	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

func (s *Service) Likes(ctx context.Context, page, pageSize, postId uint64, formatDate func(time.Time) string) (*PagedLikes, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "likes.Service.GetLiked").With("post_id", postId).With("page", page).With("page_size", pageSize)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	likesCh, eroCh := s.likesProvider.Likes(ctx, int((page-1)*pageSize), int(pageSize), postId)

	likesCount, eroErr := s.likesCountProvider.LikesNum(ctx, postId)
	if eroErr != nil {
		s.log.ErrorContext(eroErr.Context(ctx), "error while getting likes count")
		return nil, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeInternal, ErrInternal)
	}

	likes := make([]Like, 0, pageSize)
	for like := range likesCh {
		likes = append(likes, Like{
			User: User{
				Id:       like.User.Id,
				Name:     like.User.Name,
				Lastname: like.User.Lastname,
				Image:    like.User.Image,
			},
			LikedAt: formatDate(like.LikedAt),
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
		Current: page,
		Last:    (likesCount + pageSize - 1) / pageSize,
		Count:   likesCount,
		Likes:   likes,
	}, nil
}
