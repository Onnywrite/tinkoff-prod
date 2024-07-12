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
	"golang.org/x/crypto/bcrypt"
)

type UserProvider interface {
	UserByEmail(ctx context.Context, email string) (*models.User, ero.Error)
}

func PostSignIn(provider UserProvider) echo.HandlerFunc {
	type loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(c echo.Context) error {
		var data loginData
		if err := c.Bind(&data); err != nil {
			c.JSONBlob(http.StatusBadRequest, errorMessage("could not bind the body").Blob())
			return err
		}

		user, eroErr := provider.UserByEmail(context.TODO(), data.Email)
		switch {
		case errors.Is(eroErr, storage.ErrNoRows):
			c.JSONBlob(http.StatusUnauthorized, errorMessage("invalid email or password").Blob())
			return eroErr
		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
			return eroErr
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.Password)); err != nil {
			c.JSONBlob(http.StatusUnauthorized, errorMessage("invalid email or password").Blob())
			return err
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
