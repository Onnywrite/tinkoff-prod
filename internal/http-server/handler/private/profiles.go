package privatehandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/http-server/handler"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/labstack/echo/v4"
)

func GetProfile(provider handler.UserByIdProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, eroErr := provider.UserById(context.TODO(), c.Get("user_id").(uint64))
		switch {
		case errors.Is(eroErr, storage.ErrNoRows):
			c.JSONBlob(http.StatusNotFound, handler.ErrorMessage("user not found").Blob())
			return eroErr
		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, handler.ErrorMessage("internal error").Blob())
			return eroErr
		}

		if !user.IsPublic {
			return c.JSONBlob(http.StatusNotFound, handler.ErrorMessage("profile is private").Blob())
		}

		c.JSON(http.StatusOK, handler.GetProfile(user))

		return nil
	}
}
