package authhandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/http-server/handler"
	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/labstack/echo/v4"
)

type tokensResponse struct {
	Profile handler.Profile `json:"profile"`
	tokens.Pair
}

func PostRefresh(provider handler.UserByIdProvider) echo.HandlerFunc {
	type refreshToken struct {
		Refresh tokens.RefreshString `json:"refresh"`
	}

	return func(c echo.Context) error {
		var token refreshToken
		if err := c.Bind(&token); err != nil {
			c.JSONBlob(http.StatusBadRequest, handler.ErrorMessage("could not bind the body").Blob())
			return err
		}

		refresh, err := token.Refresh.ParseVerify()
		switch {
		case errors.Is(err, tokens.ErrExpired):
			c.JSONBlob(http.StatusUnauthorized, handler.ErrorMessage("refresh token has expired").Blob())
			return err
		case err != nil:
			c.JSONBlob(http.StatusUnauthorized, handler.ErrorMessage("could not parse refresh token").Blob())
			return err
		}

		user, eroErr := provider.UserById(context.TODO(), refresh.Id)
		switch {
		case errors.Is(eroErr, storage.ErrNoRows):
			c.JSONBlob(http.StatusNotFound, handler.ErrorMessage("user not found").Blob())
			return eroErr
		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, handler.ErrorMessage("internal error").Blob())
			return eroErr
		}

		pair, err := tokens.NewPair(user)
		if err != nil {
			c.JSONBlob(http.StatusInternalServerError, handler.ErrorMessage("error while generating tokens").Blob())
			return err
		}

		c.JSON(http.StatusOK, &tokensResponse{
			Profile: handler.GetProfile(user),
			Pair:    pair,
		})

		return nil
	}
}
