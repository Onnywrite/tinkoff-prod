package privatehandler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/services/feed"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/labstack/echo/v4"
)

type AllFeedProvider interface {
	AllFeed(ctx context.Context, opts feed.AllFeedOptions) (*feed.PagedFeed, ero.Error)
}

func GetFeed(provider AllFeedProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		fullTimestamp, err := strconv.ParseBool(c.QueryParam("full_timestamp"))
		if err != nil {
			fullTimestamp = false
		}
		likesCount, err := strconv.ParseUint(c.QueryParam("likes_count"), 10, 64)
		if err != nil {
			likesCount = 3
		}

		posts, eroErr := provider.AllFeed(context.Background(), feed.AllFeedOptions{
			Page:       c.Get("page").(uint64),
			PageSize:   c.Get("page_size").(uint64),
			UserId:     c.Get("id").(uint64),
			LikesCount: likesCount,
			FormatDate: func(t time.Time) string {
				if fullTimestamp {
					return t.Format(time.DateTime)
				} else {
					return t.Format(time.DateOnly)
				}
			},
		})
		switch {
		case errors.Is(eroErr, feed.ErrNoPosts):
			c.JSONBlob(http.StatusNoContent, []byte(eroErr.Error()))
			return eroErr
		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, []byte(eroErr.Error()))
			return eroErr
		}

		return c.JSON(http.StatusOK, posts)
	}
}
