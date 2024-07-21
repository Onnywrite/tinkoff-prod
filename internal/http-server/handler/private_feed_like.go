package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/services/likes"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/labstack/echo/v4"
)

type Liker interface {
	Like(ctx context.Context, userId, postId uint64) ero.Error
}

type Unliker interface {
	Unlike(ctx context.Context, userId, postId uint64) ero.Error
}

type LikesProvider interface {
	Likes(ctx context.Context, page, pageSize, postId uint64, formatDate func(time.Time) string) (*likes.PagedLikes, ero.Error)
}

func PostLike(liker Liker) echo.HandlerFunc {
	return func(c echo.Context) error {
		eroErr := liker.Like(context.TODO(), c.Get("id").(uint64), c.Get("post_id").(uint64))
		if eroErr != nil {
			return c.JSONBlob(ero.ToHttpCode(eroErr.Code()), []byte(eroErr.Error()))
		}

		c.JSONBlob(http.StatusCreated, []byte(`{}`))

		return nil
	}
}

func DeleteLike(unliker Unliker) echo.HandlerFunc {
	return func(c echo.Context) error {
		eroErr := unliker.Unlike(context.TODO(), c.Get("id").(uint64), c.Get("post_id").(uint64))
		switch {
		case errors.Is(eroErr, likes.ErrAlreadyUnliked):
			return c.JSONBlob(http.StatusTeapot, []byte(eroErr.Error()))
		}
		if eroErr != nil {
			return c.JSONBlob(ero.ToHttpCode(eroErr.Code()), []byte(eroErr.Error()))
		}

		c.JSONBlob(http.StatusCreated, []byte(`{}`))

		return nil
	}
}

func GetLikes(provider LikesProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		fullTimestamp, err := strconv.ParseBool(c.QueryParam("full_timestamp"))
		if err != nil {
			fullTimestamp = false
		}

		likesPage, eroErr := provider.Likes(context.TODO(), c.Get("page").(uint64), c.Get("page_size").(uint64), c.Get("post_id").(uint64),
			func(t time.Time) string {
				if fullTimestamp {
					return t.Format(time.DateTime)
				} else {
					return t.Format(time.DateOnly)
				}
			},
		)
		switch {
		case errors.Is(eroErr, likes.ErrNoLikes):
			c.JSONBlob(http.StatusNoContent, []byte(eroErr.Error()))
			return eroErr
		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, []byte(eroErr.Error()))
			return eroErr
		}

		c.JSON(http.StatusOK, likesPage)

		return nil
	}
}
