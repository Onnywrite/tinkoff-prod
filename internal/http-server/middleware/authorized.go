package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/labstack/echo/v4"
)

func Authorized() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header["Authorization"]
			if len(auth) == 0 {
				return c.JSONBlob(http.StatusForbidden, errorMessage("missing authorization header").Blob())
			}

			bearerToken := strings.Split(auth[0], " ")
			if bearerToken[0] != "Bearer" {
				return c.JSONBlob(http.StatusUnauthorized, errorMessage("invalid authorization header format, required 'Bearer <token>''").Blob())
			}
			access := tokens.AccessString(bearerToken[1])

			token, err := access.ParseVerify()
			switch {
			case errors.Is(err, tokens.ErrExpired):
				c.JSONBlob(http.StatusUnauthorized, errorMessage("access token has expired").Blob())
				return err
			case err != nil:
				c.JSONBlob(http.StatusUnauthorized, errorMessage("invalid token").Blob())
				return err
			}
			c.Set("email", token.Email)
			c.Set("id", token.Id)

			return next(c)
		}
	}
}
