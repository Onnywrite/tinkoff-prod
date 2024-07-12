package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/labstack/echo/v4"
)

func GetMe(provider UserProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		u, err := provider.UserByEmail(context.TODO(), c.Get("email").(string))
		switch {
		case errors.Is(err, storage.ErrNoRows):
			c.JSONBlob(http.StatusNotFound, errorMessage("user not found").Blob())
			return err
		case err != nil:
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
			return err
		}

		return c.JSON(http.StatusOK, getProfile(u))
	}
}
