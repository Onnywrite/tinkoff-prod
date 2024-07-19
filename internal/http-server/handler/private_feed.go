package handler

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
	AllFeed(ctx context.Context, page, pageSize uint64, formatDate func(time.Time) string) (*feed.PagedFeed, ero.Error)
}

func GetFeed(provider AllFeedProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		page, err := strconv.ParseUint(c.QueryParam("page"), 10, 32)
		if err != nil || page < 1 {
			page = 1
		}
		pageSize, err := strconv.ParseUint(c.QueryParam("page_size"), 10, 32)
		if err != nil || pageSize < 1 {
			pageSize = 100
		}
		fullTimestamp, err := strconv.ParseBool(c.QueryParam("full_timestamp"))
		if err != nil {
			fullTimestamp = false
		}

		posts, eroErr := provider.AllFeed(context.Background(), page, pageSize, func(t time.Time) string {
			if fullTimestamp {
				return t.Format(time.DateTime)
			} else {
				return t.Format(time.DateOnly)
			}
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
