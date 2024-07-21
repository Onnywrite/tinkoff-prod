package privatehandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/http-server/handler"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/labstack/echo/v4"
)

func GetMe(provider handler.UserProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		u, err := provider.UserByEmail(context.TODO(), c.Get("email").(string))
		switch {
		case errors.Is(err, storage.ErrNoRows):
			c.JSONBlob(http.StatusNotFound, handler.ErrorMessage("user not found").Blob())
			return err
		case err != nil:
			c.JSONBlob(http.StatusInternalServerError, handler.ErrorMessage("internal error").Blob())
			return err
		}

		return c.JSON(http.StatusOK, handler.GetProfile(u))
	}
}
