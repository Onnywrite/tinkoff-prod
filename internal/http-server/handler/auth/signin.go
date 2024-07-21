package authhandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/http-server/handler"
	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func PostSignIn(provider handler.UserProvider) echo.HandlerFunc {
	type loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(c echo.Context) error {
		var data loginData
		if err := c.Bind(&data); err != nil {
			c.JSONBlob(http.StatusBadRequest, handler.ErrorMessage("could not bind the body").Blob())
			return err
		}

		user, eroErr := provider.UserByEmail(context.TODO(), data.Email)
		switch {
		case errors.Is(eroErr, storage.ErrNoRows):
			c.JSONBlob(http.StatusUnauthorized, handler.ErrorMessage("invalid email or password").Blob())
			return eroErr
		case eroErr != nil:
			c.JSONBlob(http.StatusInternalServerError, handler.ErrorMessage("internal error").Blob())
			return eroErr
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.Password)); err != nil {
			c.JSONBlob(http.StatusUnauthorized, handler.ErrorMessage("invalid email or password").Blob())
			return err
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
