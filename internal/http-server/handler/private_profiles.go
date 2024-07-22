package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/labstack/echo/v4"
)

func GetProfile(provider UserByIdProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, eroErr := provider.UserById(context.TODO(), c.Get("user_id").(uint64))
		switch {
		case errors.Is(eroErr, storage.ErrNoRows):
			c.JSONBlob(http.StatusNotFound, errorMessage("user not found").Blob())
			return eroErr
		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
			return eroErr
		}

		if !user.IsPublic {
			return c.JSONBlob(http.StatusNotFound, errorMessage("profile is private").Blob())
		}

		c.JSON(http.StatusOK, getProfile(user))

		return nil
	}
}
