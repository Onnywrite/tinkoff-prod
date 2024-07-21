package privatehandler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Onnywrite/tinkoff-prod/internal/http-server/handler"
	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/services/feed"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/labstack/echo/v4"
)

type PostCreator interface {
	CreatePost(ctx context.Context, post feed.NewPost) (uint64, ero.Error)
}

func PostMeFeed(creator PostCreator) echo.HandlerFunc {
	type post struct {
		Content    *string  `json:"content"`
		ImagesUrls []string `json:"images_urls"`
	}

	return func(c echo.Context) error {
		var p post
		if err := c.Bind(&p); err != nil {
			c.JSONBlob(http.StatusBadRequest, handler.ErrorMessage("could not bind the body").Blob())
			return err
		}

		postId, eroErr := creator.CreatePost(context.TODO(), feed.NewPost{
			AuthorId:   c.Get("id").(uint64),
			Content:    p.Content,
			ImagesUrls: models.StringSlice(p.ImagesUrls),
		})
		switch {
		case errors.Is(eroErr, feed.ErrAuthorNotFound):
			c.JSONBlob(http.StatusNotFound, []byte(eroErr.Error()))
			return eroErr
		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, handler.ErrorMessage("internal error").Blob())
			return eroErr
		}

		return c.JSONBlob(http.StatusCreated, []byte(`{"id":`+strconv.FormatUint(postId, 10)+`}`))
	}
}
