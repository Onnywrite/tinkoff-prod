package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/services/feed"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/labstack/echo/v4"
)

type AuthorFeedProvider interface {
	AuthorFeed(ctx context.Context, page, pageSize uint64, userId uint64, formatDate func(time.Time) string) (*feed.PagedProfileFeed, ero.Error)
}

func GetProfileFeed(provider AuthorFeedProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		idStr := strings.TrimPrefix(c.Param("id"), "id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSONBlob(http.StatusNotFound, errorMessage("id is not integer").Blob())
			return err
		}
		fullTimestamp, err := strconv.ParseBool(c.QueryParam("full_timestamp"))
		if err != nil {
			fullTimestamp = false
		}

		posts, eroErr := provider.AuthorFeed(context.Background(), c.Get("page").(uint64), c.Get("page_size").(uint64), id, func(t time.Time) string {
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
