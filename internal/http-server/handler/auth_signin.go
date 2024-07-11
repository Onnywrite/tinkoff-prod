package handler

import (
	"context"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserProvider interface {
	UserByEmail(ctx context.Context, email string) (*models.User, error)
}

func PostSignIn(provider UserProvider) echo.HandlerFunc {
	type loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(c echo.Context) error {
		var data loginData
		if err := c.Bind(&data); err != nil {
			c.JSON(http.StatusUnauthorized, &crush{
				Reason: "could not bind the body",
			})
			return err
		}

		user, err := provider.UserByEmail(context.TODO(), data.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, &crush{
				Reason: "invalid email or password",
			})
			return err
		}

		if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, &crush{
				Reason: "invalid email or password",
			})
			return err
		}

		pair, err := tokens.NewPair(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &crush{
				Reason: "error while generating tokens",
			})
			return err
		}

		c.JSON(http.StatusOK, &tokensResponse{
			Profile: getProfile(user),
			Pair:    pair,
		})

		return nil
	}
}
