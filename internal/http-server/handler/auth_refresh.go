package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/labstack/echo/v4"
)

type UserByIdProvider interface {
	UserById(ctx context.Context, id uint64) (*models.User, ero.Error)
}

func PostRefresh(provider UserByIdProvider) echo.HandlerFunc {
	type refreshToken struct {
		Refresh tokens.RefreshString `json:"refresh"`
	}

	return func(c echo.Context) error {
		var token refreshToken
		if err := c.Bind(&token); err != nil {
			c.JSONBlob(http.StatusBadRequest, errorMessage("could not bind the body").Blob())
			return err
		}

		refresh, err := token.Refresh.ParseVerify()
		switch {
		case errors.Is(err, tokens.ErrExpired):
			c.JSONBlob(http.StatusUnauthorized, errorMessage("refresh token has expired").Blob())
			return err
		case err != nil:
			c.JSONBlob(http.StatusUnauthorized, errorMessage("could not parse refresh token").Blob())
			return err
		}

		user, eroErr := provider.UserById(context.TODO(), refresh.Id)
		switch {
		case errors.Is(eroErr, storage.ErrNoRows):
			c.JSONBlob(http.StatusNotFound, errorMessage("user not found").Blob())
			return eroErr
		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
			return eroErr
		}

		pair, err := tokens.NewPair(user)
		if err != nil {
			c.JSONBlob(http.StatusInternalServerError, errorMessage("error while generating tokens").Blob())
			return err
		}

		c.JSON(http.StatusOK, &tokensResponse{
			Profile: getProfile(user),
			Pair:    pair,
		})

		return nil
	}
}
