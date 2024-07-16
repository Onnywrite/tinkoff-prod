package handler

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
	"github.com/labstack/echo/v4"
)

type PostSaver interface {
	SavePost(ctx context.Context, post *models.Post) (uint64, ero.Error)
}

func PostMeFeed(poster PostSaver) echo.HandlerFunc {
	type post struct {
		Content    *string  `json:"content"`
		ImagesUrls []string `json:"images_urls"`
	}

	spaceRegex := regexp.MustCompile(`^\s*$`)
	urlRegex := regexp.MustCompile(`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b\/([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`)
	validatePost := func(p *post) ero.Error {
		type fieldFault struct {
			Field   string
			Message string
		}

		faults := make([]fieldFault, 0, 2)

		if p.Content == nil || spaceRegex.MatchString(*p.Content) {
			faults = append(faults, fieldFault{
				Field:   "content",
				Message: "cannot be empty",
			})
		} else {
			*p.Content = strings.TrimSpace(*p.Content)
		}

		if p.ImagesUrls != nil {
			for i := range p.ImagesUrls {
				p.ImagesUrls[i] = strings.Trim(p.ImagesUrls[i], " ")
				if !urlRegex.MatchString(p.ImagesUrls[i]) {
					faults = append(faults, fieldFault{
						Field:   "images_urls",
						Message: "not all URLs are valid",
					})
					break
				}
			}
			if len(p.ImagesUrls) == 0 {
				p.ImagesUrls = nil
			}
		}

		if len(faults) > 0 {
			fields := make([]string, len(faults))
			for i := range faults {
				fields[i] = faults[i].Field
			}
			return ero.NewValidation(erolog.NewContextBuilder().With("fields", fields).Build(), faults)
		}
		return nil
	}

	return func(c echo.Context) error {
		var p post
		if err := c.Bind(&p); err != nil {
			c.JSONBlob(http.StatusBadRequest, errorMessage("could not bind the body").Blob())
			return err
		}

		if eroErr := validatePost(&p); eroErr != nil {
			c.JSON(http.StatusBadRequest, eroErr)
			return eroErr
		}

		postId, eroErr := poster.SavePost(context.TODO(), &models.Post{
			Author: models.User{
				Id: c.Get("id").(uint64),
			},
			Content:    *p.Content,
			ImagesUrls: models.StringSlice(p.ImagesUrls),
		})
		switch {
		case errors.Is(eroErr, storage.ErrForeignKeyConstraint):
			c.JSONBlob(http.StatusNotFound, errorMessage("user (author) not found").Blob())
			return eroErr
		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
			return eroErr
		}

		return c.JSONBlob(http.StatusCreated, []byte(`{"id":`+strconv.FormatUint(postId, 10)+`}`))
	}
}
