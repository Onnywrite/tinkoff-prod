package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/labstack/echo/v4"
)

type PostsProvider interface {
	Posts(ctx context.Context, offset, count int) ([]models.Post, ero.Error)
}

type PostsCountProvider interface {
	PostsNum(context.Context) (uint64, ero.Error)
}

func GetFeed(provider PostsProvider, countProvider PostsCountProvider) echo.HandlerFunc {
	type author struct {
		Id       uint64 `json:"id"`
		Name     string `json:"name"`
		Lastname string `json:"surname"`
	}
	type post struct {
		Id          uint64  `json:"id"`
		Author      author  `json:"author"`
		Content     string  `json:"content"`
		ImageUrl    *string `json:"image_url"`
		PublishedAt string  `json:"published_at"`
		UpdatedAt   *string `json:"updated_at"`
	}
	type response struct {
		First   uint64 `json:"first"`
		Current uint64 `json:"current"`
		Last    uint64 `json:"last"`
		Posts   []post `json:"posts"`
	}

	formatTime := func(t time.Time, fullTimestamp bool) string {
		if fullTimestamp {
			return t.Format(time.DateTime)
		} else {
			return t.Format(time.DateOnly)
		}
	}

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

		posts, eroErr := provider.Posts(context.TODO(), int(page-1)*int(pageSize), int(pageSize))
		switch {
		case errors.Is(eroErr, storage.ErrNoRows):
			c.JSONBlob(http.StatusNoContent, errorMessage("no content").Blob())
			return eroErr

		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
			return eroErr
		}

		postsCount, eroErr := countProvider.PostsNum(context.TODO())
		if eroErr != nil {
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
			return eroErr
		}

		newPosts := make([]post, len(posts))
		for i, p := range posts {
			var url *string
			if p.ImagesUrls == nil || len(p.ImagesUrls) == 0 {
				url = nil
			} else {
				url = &p.ImagesUrls[0]
			}

			var updatedAt *string
			if p.UpdatedAt != nil {
				formatted := formatTime(*p.UpdatedAt, fullTimestamp)
				updatedAt = &formatted
			} else {
				updatedAt = nil
			}

			newPosts[i] = post{
				Id:          p.Id,
				Content:     p.Content,
				ImageUrl:    url,
				PublishedAt: formatTime(p.PublishedAt, fullTimestamp),
				UpdatedAt:   updatedAt,
				Author: author{
					Id:       p.Author.Id,
					Name:     p.Author.Name,
					Lastname: p.Author.Lastname,
				},
			}
		}

		return c.JSON(http.StatusOK, response{
			First:   1,
			Current: uint64(page),
			Last:    (postsCount + pageSize + 1) / pageSize,
			Posts:   newPosts,
		})
	}
}
