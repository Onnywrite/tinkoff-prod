package middleware

import (
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
				return c.JSON(http.StatusUnauthorized, &crush{
					Reason: "unauthorized user",
				})
			}

			bearerToken := strings.Split(auth[0], " ")
			if bearerToken[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, &crush{
					Reason: "invalid authorization header format",
				})
			}
			access := tokens.AccessString(bearerToken[1])

			token, err := access.ParseVerify()
			if err != nil {
				c.JSON(http.StatusUnauthorized, &crush{
					Reason: "invalid token",
				})
				return err
			}
			c.Set("email", token.Email)
			c.Set("id", token.Id)

			return next(c)
		}
	}
}
